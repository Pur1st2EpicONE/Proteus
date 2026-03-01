(function () {
  const copyIdBtn = document.getElementById("copyIdBtn");
  const form = document.getElementById("uploadForm");
  const imageInput = document.getElementById("imageInput");
  const actionSelect = document.getElementById("actionSelect");
  const widthInput = document.getElementById("widthInput");
  const heightInput = document.getElementById("heightInput");
  const watermarkInput = document.getElementById("watermarkInput");

  const widthField = document.getElementById("widthField");
  const heightField = document.getElementById("heightField");
  const watermarkField = document.getElementById("watermarkField");

  const statusBlock = document.getElementById("statusBlock");
  const statusText = document.getElementById("statusText");

  const previewBlock = document.getElementById("previewBlock");
  const resultImage = document.getElementById("resultImage");
  const imageIdNode = document.getElementById("imageId");
  const downloadBtn = document.getElementById("downloadBtn");
  const deleteBtn = document.getElementById("deleteBtn");
  const clearBtn = document.getElementById("clearBtn");

  const searchIdInput = document.getElementById("searchIdInput");
  const searchBtn = document.getElementById("searchBtn");

  let currentId = null;
  let pollTimer = null;

  function resetUi() {
    statusBlock.hidden = true;
    previewBlock.hidden = true;
    statusText.textContent = "";
    resultImage.src = "";
    imageIdNode.textContent = "";
    currentId = null;
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  function updateFieldsVisibility() {
    const action = actionSelect.value;
    const needsResize = action === "resize";
    const needsWatermark = action === "watermark";

    widthField.classList.toggle("hidden", !needsResize);
    heightField.classList.toggle("hidden", !needsResize);
    watermarkField.classList.toggle("hidden", !needsWatermark);
  }

  clearBtn.addEventListener("click", () => {
    form.reset();
    resetUi();
  });

  actionSelect.addEventListener("change", updateFieldsVisibility);

  form.addEventListener("submit", async (ev) => {
    ev.preventDefault();
    if (!imageInput.files || imageInput.files.length === 0) {
      return alert("Please select a file");
    }

    const action = actionSelect.value;
    const needsWatermark = action === "watermark";

    if (needsWatermark && !watermarkInput.value.trim()) {
      return alert("Watermark text is required for this action");
    }

    const fd = new FormData();
    fd.append("image", imageInput.files[0]);
    fd.append("action", action);
    if (widthInput.value) fd.append("width", widthInput.value);
    if (heightInput.value) fd.append("height", heightInput.value);
    if (watermarkInput.value) fd.append("watermark", watermarkInput.value);

    statusBlock.hidden = false;
    statusText.textContent = "Uploading...";

    try {
      const res = await fetch("/api/v1/upload", { method: "POST", body: fd });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        statusText.textContent = `Upload error: ${res.status} ${err.error || ""}`;
        return;
      }

      const body = await res.json();
      const id = body.result;
      currentId = id;
      imageIdNode.textContent = id;

      statusText.textContent = "In queue for processing...";
      startPolling(id);
    } catch (err) {
      statusText.textContent = "Network error: " + err.message;
    }
  });

  function startPolling(id) {
    let elapsed = 0;
    const interval = 2000;
    const timeout = 60000;

    if (pollTimer) clearInterval(pollTimer);

    pollTimer = setInterval(async () => {
      elapsed += interval;
      try {
        const res = await fetch(`/api/v1/image/${encodeURIComponent(id)}`);

        if (res.status === 202) {
          statusText.textContent = "Processing...";
          return;
        }

        if (res.status === 200) {
          const blob = await res.blob();
          const url = URL.createObjectURL(blob);
          resultImage.src = url;
          previewBlock.hidden = false;
          imageIdNode.textContent = id;
          currentId = id;
          setupPreviewControls(id, url);

          statusBlock.hidden = true;
          clearInterval(pollTimer);
          pollTimer = null;
          return;
        }

        if (res.status === 404) {
          statusText.textContent = "Image not found (404)";
          clearInterval(pollTimer);
          pollTimer = null;
          return;
        }

        const txt = await res.text();
        statusText.textContent = `Error: ${res.status} ${txt}`;
        clearInterval(pollTimer);
        pollTimer = null;
      } catch (e) {
        statusText.textContent = "Network error while polling: " + e.message;
        clearInterval(pollTimer);
        pollTimer = null;
      }

      if (elapsed >= timeout) {
        statusText.textContent = "Timeout waiting for result.";
        clearInterval(pollTimer);
        pollTimer = null;
      }
    }, interval);
  }

  function setupPreviewControls(id, blobUrl) {
    downloadBtn.onclick = () => {
      const a = document.createElement("a");
      a.href = blobUrl;
      a.download = `${id}.jpg`;
      document.body.appendChild(a);
      a.click();
      a.remove();
    };

    deleteBtn.onclick = async () => {
      if (!confirm("Delete the image on the server?")) return;
      try {
        const res = await fetch(`/api/v1/image/${encodeURIComponent(id)}`, {
          method: "DELETE",
        });
        if (res.ok) {
          resetUi();
          statusBlock.hidden = false;
          statusText.textContent = "Image deleted";
        } else {
          const err = await res.json().catch(() => ({}));
          alert("Delete error: " + (err.error || res.status));
        }
      } catch (e) {
        alert("Network error: " + e.message);
      }
    };

    copyIdBtn.onclick = () => {
      navigator.clipboard
        .writeText(id)
        .then(() => {
          const original = copyIdBtn.textContent;
          copyIdBtn.textContent = "🔲";
          copyIdBtn.style.color = "#10b981";
          setTimeout(() => {
            copyIdBtn.textContent = original;
            copyIdBtn.style.color = "";
          }, 1500);
        })
        .catch(() => {
          alert("Failed to copy to clipboard");
        });
    };
  }

  searchBtn.addEventListener("click", async () => {
    const id = searchIdInput.value.trim();
    if (!id) return alert("Enter image ID");

    resetUi();
    statusBlock.hidden = false;
    statusText.textContent = "Loading...";

    try {
      const res = await fetch(`/api/v1/image/${encodeURIComponent(id)}`);

      if (res.status === 202) {
        currentId = id;
        imageIdNode.textContent = id;
        statusText.textContent = "In queue for processing...";
        startPolling(id);
        return;
      }

      if (res.status === 200) {
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        resultImage.src = url;
        previewBlock.hidden = false;
        imageIdNode.textContent = id;
        currentId = id;
        setupPreviewControls(id, url);

        statusBlock.hidden = true;
        return;
      }

      if (res.status === 404) {
        statusText.textContent = "Image not found (404)";
        return;
      }

      const txt = await res.text();
      statusText.textContent = `Error: ${res.status} ${txt}`;
    } catch (e) {
      statusText.textContent = "Network error: " + e.message;
    }
  });

  updateFieldsVisibility();
})();
