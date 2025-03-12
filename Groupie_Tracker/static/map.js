document.addEventListener("DOMContentLoaded", function () {
    var map = L.map('map').setView([48.8566, 2.3522], 5);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors'
    }).addTo(map);

    var locations = JSON.parse(document.getElementById("locations-data").textContent);

    locations.forEach(function(loc) {
        fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(loc.city)}`)
            .then(response => response.json())
            .then(data => {
                if (data.length > 0) {
                    var lat = data[0].lat;
                    var lon = data[0].lon;

                    var marker = L.marker([lat, lon]).addTo(map);
                    marker.bindPopup(`<b>${loc.city}</b><br>${loc.dates}`);
                }
            })
            .catch(error => console.error("Erreur chargement lieu :", error));
    });
});
