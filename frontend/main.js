var map;
currentFeature_or_Features = null;
var stateHouseLatitude = 39.768732;
var stateHouseLongitude = -86.162612;

var geojson_MultiPolygon = {
    "type": "MultiPolygon",
    "coordinates": [
        [
            [
                [
                    -86.157712,
                    39.776782
                ],
                [
                    -86.157671,
                    39.777967
                ],
                [
                    -86.157669,
                    39.777994
                ],
                [
                    -86.15763,
                    39.777998
                ],
                [
                    -86.157594,
                    39.777998
                ],
                [
                    -86.156763,
                    39.777978
                ],
                [
                    -86.155858,
                    39.777954
                ],
                [
                    -86.155858,
                    39.777891
                ],
                [
                    -86.155878,
                    39.777436
                ],
                [
                    -86.155905,
                    39.776742
                ],
                [
                    -86.155922,
                    39.776328
                ],
                [
                    -86.155955,
                    39.775468
                ],
                [
                    -86.157759,
                    39.775508
                ],
                [
                    -86.157712,
                    39.776782
                ]
            ]
        ],
        [
            [
                [
                    -86.155905,
                    39.776742
                ],
                [
                    -86.154999,
                    39.776731
                ],
                [
                    -86.154092,
                    39.776702
                ],
                [
                    -86.154217,
                    39.776671
                ],
                [
                    -86.154256,
                    39.776663
                ],
                [
                    -86.154276,
                    39.776654
                ],
                [
                    -86.15431,
                    39.776636
                ],
                [
                    -86.15433,
                    39.776623
                ],
                [
                    -86.154737,
                    39.776301
                ],
                [
                    -86.155922,
                    39.776328
                ],
                [
                    -86.155905,
                    39.776742
                ]
            ]
        ]
    ]
};

var roadStyle = {
    strokeColor: "#FFFF00",
    strokeWeight: 1,
    strokeOpacity: 0.75
};

var parcelStyle = {
    strokeColor: "#FF7800",
    strokeOpacity: 1,
    strokeWeight: 1,
    fillColor: "#46461F",
    fillOpacity: 0.25
};

var infowindow = new google.maps.InfoWindow();

function init() {
    map = new google.maps.Map(document.getElementById('map'), {
        zoom: 14,
        mapTypeId: google.maps.MapTypeId.ROADMAP
    });
    if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
            function(position) {
                var pos = new google.maps.LatLng(
                    position.coords.latitude,
                    position.coords.longitude
                );

                var infowindow = new google.maps.InfoWindow({});

                map.setCenter(pos);
            },
            function() {
                map.setCenter(new google.maps.LatLng(
                    stateHouseLatitude, stateHouseLongitude
                ));
            }
        );
    } else {
        // Browser doesn't support Geolocation
        map.setCenter(new google.maps.LatLng(
            stateHouseLatitude, stateHouseLongitude
        ));
    }
}

function clearMap() {
    if (!currentFeature_or_Features)
        return;
    if (currentFeature_or_Features.length) {
        for (var i = 0; i < currentFeature_or_Features.length; i++) {
            if (currentFeature_or_Features[i].length) {
                for (var j = 0; j < currentFeature_or_Features[i].length; j++) {
                    currentFeature_or_Features[i][j].setMap(null);
                }
            }
            else {
                currentFeature_or_Features[i].setMap(null);
            }
        }
    } else {
        currentFeature_or_Features.setMap(null);
    }
    if (infowindow.getMap()) {
        infowindow.close();
    }
}
function showFeature(geojson, style) {
    clearMap();
    currentFeature_or_Features = new GeoJSON(geojson, style || null);
    if (currentFeature_or_Features.type && currentFeature_or_Features.type == "Error") {
        document.getElementById("put_geojson_string_here").value = currentFeature_or_Features.message;
        return;
    }
    if (currentFeature_or_Features.length) {
        for (var i = 0; i < currentFeature_or_Features.length; i++) {
            if (currentFeature_or_Features[i].length) {
                for (var j = 0; j < currentFeature_or_Features[i].length; j++) {
                    currentFeature_or_Features[i][j].setMap(map);
                    if (currentFeature_or_Features[i][j].geojsonProperties) {
                        setInfoWindow(currentFeature_or_Features[i][j]);
                    }
                }
            }
            else {
                currentFeature_or_Features[i].setMap(map);
            }
            if (currentFeature_or_Features[i].geojsonProperties) {
                setInfoWindow(currentFeature_or_Features[i]);
            }
        }
    } else {
        currentFeature_or_Features.setMap(map)
        if (currentFeature_or_Features.geojsonProperties) {
            setInfoWindow(currentFeature_or_Features);
        }
    }

    document.getElementById("put_geojson_string_here").value = JSON.stringify(geojson);
}
function rawGeoJSON() {
    clearMap();
    currentFeature_or_Features = new GeoJSON(JSON.parse(document.getElementById("put_geojson_string_here").value));
    if (currentFeature_or_Features.length) {
        for (var i = 0; i < currentFeature_or_Features.length; i++) {
            if (currentFeature_or_Features[i].length) {
                for (var j = 0; j < currentFeature_or_Features[i].length; j++) {
                    currentFeature_or_Features[i][j].setMap(map);
                    if (currentFeature_or_Features[i][j].geojsonProperties) {
                        setInfoWindow(currentFeature_or_Features[i][j]);
                    }
                }
            }
            else {
                currentFeature_or_Features[i].setMap(map);
            }
            if (currentFeature_or_Features[i].geojsonProperties) {
                setInfoWindow(currentFeature_or_Features[i]);
            }
        }
    } else {
        currentFeature_or_Features.setMap(map);
        if (currentFeature_or_Features.geojsonProperties) {
            setInfoWindow(currentFeature_or_Features);
        }
    }
}
function setInfoWindow(feature) {
    google.maps.event.addListener(feature, "click", function(event) {
        var content = "<div id='infoBox'><strong>GeoJSON Feature Properties</strong><br />";
        for (var j in this.geojsonProperties) {
            content += j + ": " + this.geojsonProperties[j] + "<br />";
        }
        content += "</div>";
        infowindow.setContent(content);
        infowindow.setPosition(event.latLng);
        infowindow.open(map);
    });
}