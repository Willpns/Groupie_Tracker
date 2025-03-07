// =============== Existing sorting code ===============
document.querySelectorAll(".sort-menu a").forEach(link => {
    link.addEventListener("click", function (event) {
        event.preventDefault(); // Prevent full page reload

        const url = this.href; // e.g. /home?sortBy=name&order=asc
        fetch(url, { headers: { "X-Requested-With": "XMLHttpRequest" } })
            .then(response => response.text())
            .then(html => {
                // Replace only the .artists section
                const parsed = new DOMParser().parseFromString(html, "text/html");
                const newArtists = parsed.querySelector(".artists");
                if (newArtists) {
                    document.querySelector(".artists").innerHTML = newArtists.innerHTML;
                }
            })
            .catch(error => console.error("Sorting failed:", error));
    });
});

// =============== NEW: Filter form handling ===============
const filterForm = document.getElementById("filter-form");
if (filterForm) {
    filterForm.addEventListener("submit", function(event) {
        event.preventDefault(); // Don’t do a full-page submit

        // Gather all fields from the filter form
        const formData = new FormData(filterForm);
        // Turn them into a URL query string
        const queryString = new URLSearchParams(formData).toString();
        // We’ll send it to /home just like sorting
        const url = "/home?" + queryString;

        fetch(url, { headers: { "X-Requested-With": "XMLHttpRequest" } })
            .then(response => response.text())
            .then(html => {
                // Replace the .artists container with the filtered results
                const parsed = new DOMParser().parseFromString(html, "text/html");
                const newArtists = parsed.querySelector(".artists");
                if (newArtists) {
                    document.querySelector(".artists").innerHTML = newArtists.innerHTML;
                }
            })
            .catch(error => console.error("Filtering failed:", error));
    });
}
