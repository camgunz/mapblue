var TileJSONs = [
    'https://a.tiles.mapbox.com/v3/examples.map-20v6611k,mapbox.dc-property-values.jsonp?secure',
];

$('#map').mapbox(TileJSONs, function(map) {

    stateHouseLatitude = 39.768732;
    stateHouseLongitude = -86.162612;

    // Set initial latitude, longitude and zoom level
    map.setCenterZoom({
        lat: stateHouseLatitude,
        lon: stateHouseLongitude
    }, 16);

    // Set minimum and maximum zoom levels
    map.setZoomRange(15, 18);
/*
    map.data.setStyle(styleBlock);

    google.maps.event.addListener(map, 'tilesloaded', updateBlocks);
    google.maps.event.addListener(map, 'zoom_changed', updateBlocks);
    google.maps.event.addListener(map, 'dragend', updateBlocks);
    */

});

var blackCoeff = 0.4501479;
var hispanicCoeff = 0.077551;
var otherRaceCoeff = 0.1358834;
var unmarriedCoeff = 0.0911239;
var childlessCoeff = 0.115441;
var regressionConstant = 0.3638054;
var mapblueAPI = "http://totaltrash.org/mapblue/lookup"
var stateHouseLatitude = 39.768732;
var stateHouseLongitude = -86.162612;

function getDemProbability(over18, black, hispanic, otherRace, unmarried,
        childless) {
    if (over18 == 0) {
        return 0;
    }

    return regressionConstant +
            (blackCoeff * (black / over18)) +
            (hispanicCoeff * (hispanic / over18)) +
            (otherRaceCoeff * (otherRace / over18)) +
            (unmarriedCoeff * (unmarried / over18)) +
            (childlessCoeff * (childless / over18));
}

function buildAPIURL(lat1, lon1, lat2, lon2) {
    return mapblueAPI + "?lat1=" + lat1 +
            "&lon1=" + lon1 +
            "&lat2=" + lat2 +
            "&lon2=" + lon2;
}

function styleBlock(block) {
    var over18 = block.getProperty('over18');
    var black = block.getProperty('black');
    var hispanic = block.getProperty('hispanic');
    var otherRace = block.getProperty('otherRace');
    var unmarried = block.getProperty('unmarried');
    var childless = block.getProperty('childless');
    var demProbability = getDemProbability(
            block.getProperty('over18'),
            block.getProperty('black'),
            block.getProperty('hispanic'),
            block.getProperty('otherRace'),
            block.getProperty('unmarried'),
            block.getProperty('childless')
            )

    if (over18 < .3) {
        return {
            fillColor: "#BBBBBB",
            fillOpacity: .5,
            strokeWeight: 1
        };
    }

    if (demProbability < .5) {
        return {
            fillColor: "#BB4444",
            fillOpacity: demProbability,
            strokeWeight: 1
        };
    }

    return {
        fillColor: "#4488CC",
        fillOpacity: demProbability,
        strokeWeight: 1
    };
}

function restyleBlocks() {

    blackCoeff = parseFloat(document.getElementById('blackCoeff').value);
    hispanicCoeff = parseFloat(document.getElementById('hispanicCoeff').value);
    otherRaceCoeff = parseFloat(document.getElementById('otherRaceCoeff').value);
    unmarriedCoeff = parseFloat(document.getElementById('unmarriedCoeff').value);
    childlessCoeff = parseFloat(document.getElementById('childlessCoeff').value);
    regressionConstant = parseFloat(document.getElementById('regressionConstant').value);

    map.data.forEach(function(feature) {
        map.data.overrideStyle(feature, styleBlock(feature));
    });
}

function updateBlocks() {
    var bounds = map.getBounds();
    var northEast = bounds.getNorthEast();
    var southWest = bounds.getSouthWest();

    map.data.loadGeoJson(buildAPIURL(
            northEast.lat(), northEast.lng(), southWest.lat(), southWest.lng()
            ))
}