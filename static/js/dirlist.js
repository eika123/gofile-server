document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll("li.file[data-name]").forEach((item) => {
    const name = item.dataset.name || "";
    const ext = name.split(".").pop().toLowerCase();

    const category = getFileCategory(ext);
    item.classList.add(`type-${category}`);
  });
});

function getFileCategory(extension) {
  switch (extension) {
  case "pdf":
    return "pdf";
  case "jpg":
  case "jpeg":
  case "png":
  case "gif":
  case "svg":
    return "image";
  case "txt":
  case "md":
  case "json":
  case "yaml":
  case "yml":
    return "text";
  case "zip":
  case "tar":
  case "gz":
  case "rar":
    return "archive";
  default:
    return "file";
  }
}
