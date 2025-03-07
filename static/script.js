document.querySelectorAll(".sort-menu a").forEach(link => {
    link.addEventListener("click", function (event) {
        event.preventDefault(); // Prevent full page reload

        const url = this.href; // Get the sorting URL
        fetch(url, { headers: { "X-Requested-With": "XMLHttpRequest" } }) // Send AJAX request
            .then(response => response.text())
            .then(html => {
                document.querySelector(".artists").innerHTML = 
                    new DOMParser().parseFromString(html, "text/html")
                                   .querySelector(".artists").innerHTML;
            })
            .catch(error => console.error("Sorting failed:", error));
    });
});