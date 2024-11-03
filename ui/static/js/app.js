const navLinks = document.querySelectorAll("nav a");
const features = document.querySelector(".features");
const featuresBtn = document.querySelector(".features-add-btn");
const featureInput = document.querySelector(".features-input");
const featuresSubmitBtn = document.querySelector(".features-continue-btn");
const uploadInput = document.querySelector("#upload-input");
const uploadBtn = document.querySelector("#upload-btn");

// handles navLink clicks
for (var i = 0; i < navLinks.length; i++) {
  var link = navLinks[i];
  if (link.getAttribute("href") == window.location.pathname) {
    link.classList.add("live");
    break;
  }
}

// handles adding and removing the user feature
if (featuresBtn && featureInput) {
  let inputText = "";
  let featuresArr = [];

  featureInput.addEventListener("input", (e) => {
    inputText = e.target.value;
  });

  featuresBtn.addEventListener("click", (e) => {
    e.preventDefault();

    if (inputText !== "" && features.childElementCount < 3) {
      const featureDiv = document.createElement("div");
      featureDiv.innerHTML = `
        ${inputText}
        <span>x</span>
      `;

      const featureDelBtn = featureDiv.querySelector("span");

      featureDelBtn.addEventListener("click", (e) => {
        if (featuresArr.length !== 0) {
          e.target.parentElement.remove();

          const elToRemove = e.target.parentElement.textContent
            .trim()
            .split("\n")[0];
          const newFeatures = featuresArr.filter((el) => el !== elToRemove);
          featuresArr = newFeatures;
        }
      });

      featuresArr.push(inputText.toLocaleLowerCase());
      features.appendChild(featureDiv);

      featureInput.value = "";
      inputText = "";
      console.log(featuresArr);
    }
  });

  featuresSubmitBtn.addEventListener("click", (e) => {
    if (featuresArr.length !== 0) {
      const formData = new FormData();

      const csrfToken = document.querySelector(
        "input[name='csrf_token']"
      ).value;

      formData.append("features", JSON.stringify(featuresArr));
      formData.append("csrf_token", csrfToken);

      fetch("/api/features", {
        method: "POST",
        body: formData,
      })
        .then((response) => {
          if (response.status === 303) {
            return response.json();
          }
        })
        .then((data) => {
          if (data.redirect) {
            window.location.href = data.redirect;
          }
        })
        .catch((error) => console.log("Error:", error));
    }
  });
}

// handles uploading company past data
if (uploadBtn && uploadInput) {
  function trimFileName(word) {
    const extensionSlice = word.length - 5;
    if (word.split(".")[0].length > 12) {
      return `${word.slice(0, 10)}...${word.slice(extensionSlice)}`;
    }

    return word;
  }

  function formatFileSize(size) {
    if (size < 1024) return `${size} B`;
    else if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`;
    else return `${(size / (1024 * 1024)).toFixed(2)} MB`;
  }

  let file = null;
  uploadBtn.addEventListener("click", () => {
    uploadInput.click();
  });

  uploadInput.addEventListener("change", () => {
    file = uploadInput.files[0];
    console.log(file);

    document.querySelector(".upload-container p").textContent = `${trimFileName(
      file.name
    )} - ${formatFileSize(file.size)}`;
  });

  featuresSubmitBtn.addEventListener("click", () => {
    const csrfToken = document.querySelector("input[name='csrf_token']").value;

    const formData = new FormData();

    formData.append("file", file);
    formData.append("csrf_token", csrfToken);

    fetch("/api/import", {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.ok) {
          console.log("File uploaded successfully");
          window.location.href = "/predict";
        } else {
          console.log("File upload failed");
        }
      })
      .catch((error) => console.log("Error:", error));
  });
}
