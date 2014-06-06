var map;
var blackCoeff = 0.4501479;
var hispanicCoeff = 0.077551;
var otherRaceCoeff = 0.1358834;
var unmarriedCoeff = 0.0911239;
var childlessCoeff = 0.115441;
var regressionConstant = 0.3638054;
var mapblueAPI = "http://totaltrash.org/mapblue/lookup"
var stateHouseLatitude = 39.768732;
var stateHouseLongitude = -86.162612;
var features = [];

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

// var infowindow = new google.maps.InfoWindow();

function getDemPercentage(over18, black, hispanic, otherRace, unmarried,
                          childless) {
    if (over18 == 0) {
        return 0;
    }

    return regressionConstant +
        (blackCoeff * (black / over18)) +
        (hispanicCoeff * (otherRace / over18)) +
        (otherRaceCoeff * (otherRace / over18)) +
        (unmarriedCoeff * (unmarried / over18)) +
        (childlessCoeff * (childless / over18));
}

function handleJSONResponse(data) {
    if (features.length > 0) {
        $.each(features, function(index, feature) {
            map.data.remove(feature);
        });
    }

    $.each(data, function(key, blocks) {
        $.each(blocks, function(index, block) {
            var newFeatures = map.data.addGeoJson(block);

            $.each(newFeatures, function(index, newFeature) {
                map.data.overrideStyle(newFeature, {
                    fillColor: "#4488CC",
                    fillOpacity: getDemPercentage(
                        block.properties.over18,
                        block.properties.black,
                        block.properties.hispanic,
                        block.properties.otherRace,
                        block.properties.unmarried,
                        block.properties.childless
                    )
                });
            });
            features = features.concat(newFeatures);
        });
    });
}

function loadBlocks() {
    var bounds = map.getBounds();
    var northEast = bounds.getNorthEast();
    var southWest = bounds.getSouthWest();

    $.getJSON(mapblueAPI, {
        lat1: northEast.lat(),
        lon1: southWest.lng(),
        lat2: southWest.lat(),
        lon2: northEast.lng()
    }, handleJSONResponse);
}

function init() {
    var latitude = stateHouseLatitude;
    var longitude = stateHouseLongitude;

    if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(function(position) {
            latitude = position.coords.latitude;
            longitude = position.coords.longitude;
        });
    }

    map = new google.maps.Map(document.getElementById('map'), {
        center: new google.maps.LatLng(latitude, longitude),
        zoom: 16
    });

    google.maps.event.addListener(map, 'bounds_changed', loadBlocks);
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