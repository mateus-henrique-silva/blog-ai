/* AI Studies — Markdown Editor Init */
(function () {
  var textarea = document.getElementById("content_md");
  if (!textarea) return;

  var easyMDE = new EasyMDE({
    element: textarea,
    spellChecker: false,
    autosave: {
      enabled: true,
      uniqueId: "post-editor-" + (window.location.pathname),
      delay: 3000,
    },
    minHeight: "400px",
    placeholder: "Write your post in Markdown...",
    toolbar: [
      "bold", "italic", "strikethrough", "heading", "|",
      "quote", "unordered-list", "ordered-list", "horizontal-rule", "|",
      "link", "|",
      {
        name: "upload-media",
        action: uploadMedia,
        className: "fa fa-upload",
        title: "Upload Image / Video / Audio",
        attributes: { id: "upload-media-btn" },
      },
      "|",
      "preview", "side-by-side", "fullscreen", "|",
      "guide",
    ],
    previewRender: function (text) {
      return EasyMDE.prototype.markdown(text);
    },
  });

  function uploadMedia() {
    var input = document.createElement("input");
    input.type = "file";
    input.accept = "image/*,video/mp4,video/webm,audio/mpeg,audio/ogg";

    input.onchange = function () {
      var file = input.files && input.files[0];
      if (!file) return;

      var formData = new FormData();
      formData.append("file", file);

      fetch("/studio/upload", {
        method: "POST",
        body: formData,
      })
        .then(function (res) {
          if (!res.ok) {
            return res.json().then(function (d) {
              throw new Error(d.error || "Upload failed");
            });
          }
          return res.json();
        })
        .then(function (data) {
          var md = "";
          if (data.mime_type.indexOf("image/") === 0) {
            md = "![" + (data.filename || file.name) + "](" + data.url + ")";
          } else if (data.mime_type.indexOf("video/") === 0) {
            md =
              '<video controls src="' +
              data.url +
              '" style="max-width:100%"></video>';
          } else if (data.mime_type.indexOf("audio/") === 0) {
            md = '<audio controls src="' + data.url + '"></audio>';
          }
          var cm = easyMDE.codemirror;
          cm.replaceSelection(md);
        })
        .catch(function (err) {
          alert("Upload failed: " + err.message);
        });
    };

    input.click();
  }
})();
