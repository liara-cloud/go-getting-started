const fileInput = document.getElementById('fileInput');
const preview = document.getElementById('preview');
const message = document.getElementById('message');

fileInput.addEventListener('change', () => {
    const file = fileInput.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = (e) => {
            preview.innerHTML = `<img src="${e.target.result}" alt="Preview">`;
        };
        reader.readAsDataURL(file);
        uploadFile(file);
    }
});

function uploadFile(file) {
    const formData = new FormData();
    formData.append('image', file);

    fetch('/upload', {
        method: 'POST',
        body: formData,
    })
        .then((response) => response.json())
        .then((data) => {
            if (data.url) {
                message.innerHTML = `<span style="color: green;">Uploaded successfully: <a href="${data.url}" target="_blank">View Image</a></span>`;
            } else {
                message.innerHTML = `<span style="color: red;">Upload failed: ${data.message}</span>`;
            }
        })
        .catch((err) => {
            message.innerHTML = `<span style="color: red;">Error: ${err.message}</span>`;
        });
}
