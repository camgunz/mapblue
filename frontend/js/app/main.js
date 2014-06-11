var map = null;
var blackCoeff = 0.4501479;
var hispanicCoeff = 0.077551;
var otherRaceCoeff = 0.1358834;
var unmarriedCoeff = 0.0911239;
var childlessCoeff = 0.115441;
var regressionConstant = 0.3638054;
var mapblueAPI = 'http://mapblue.org/lookup'
var stateHouseLatitude = 39.768732;
var stateHouseLongitude = -86.162612;
var mapboxTileJSON = 'https://a.tiles.mapbox.com/v3/examples.map-20v6611k,mapbox.dc-property-values.jsonp?secure';
var loadedBlocks = [];

function getDemProbability(over18, black, hispanic, otherRace, unmarried,
                           childless) {
    if (over18 == 0) {
        return 0;
    }

    return regressionConstant + (blackCoeff * (black / over18)) +
                                (hispanicCoeff * (hispanic / over18)) +
                                (otherRaceCoeff * (otherRace / over18)) +
                                (unmarriedCoeff * (unmarried / over18)) +
                                (childlessCoeff * (childless / over18));
}

function getBlockDemProbability(block) {
    return getDemProbability(
        block.properties.over18,
        block.properties.black,
        block.properties.hispanic,
        block.properties.otherRace,
        block.properties.unmarried,
        block.properties.childless
    );
}

function buildAPIURL(lat1, lon1, lat2, lon2) {
    return mapblueAPI + '?lat1=' + lat1 +
                        '&lon1=' + lon1 +
                        '&lat2=' + lat2 +
                        '&lon2=' + lon2;
}

function getMapCoordinates() {
    var bounds = map.getBounds();
    var northEast = bounds.getNorthEast();
    var southWest = bounds.getSouthWest();

    return {
        lat1: northEast.lat,
        lon1: northEast.lng,
        lat2: southWest.lat,
        lon2: southWest.lng
    };
}

function blockStyler(block) {
    var demProbability = getBlockDemProbability(block);

    if (block.properties.over18 == 0) {
        return {
            fillColor: '#FFFFFF',
            fillOpacity: 0.0,
            stroke: false
        };
    }

    if (demProbability < .5) {
        return {
            fillColor: '#BB4444',
            fillOpacity: 1.0 - demProbability,
            stroke: false
        };
    }

    return {
        fillColor: '#4488CC',
        fillOpacity: demProbability,
        stroke: false
    };
}

function blockFilter(block) {
    if (loadedBlocks.indexOf(block.id) == -1) {
        loadedBlocks.push(block.id);
        return true;
    }
    return false;
}

function loadBlocks() {
    $.getJSON(mapblueAPI, getMapCoordinates(), function(data) {
        var geoJSON = map.featureLayer.getGeoJSON();

        if (!geoJSON) {
            L.geoJson(data, {
                style: blockStyler,
                filter: blockFilter
            }).addTo(map);
        }
        else {
            geoJSON.addData(data);
        }

        // L.geoJson(data, { style: blockStyler }).addTo(map);

        // L.geoJson(data, { style: blockStyler }).addTo(map);
        // map.featureLayer = L.geoJson(data, { style: blockStyler });
    });
}

function init() {
    // $(function() {$('body')});

    $('#blackCoeff').val(blackCoeff);
    $('#hispanicCoeff').val(hispanicCoeff);
    $('#otherRaceCoeff').val(otherRaceCoeff);
    $('#unmarriedCoeff').val(unmarriedCoeff);
    $('#childlessCoeff').val(childlessCoeff);
    $('#regressionConstant').val(regressionConstant);

    map = L.mapbox.map('map', mapboxTileJSON, {
        center: [stateHouseLatitude, stateHouseLongitude],
        zoom: 16,
        minZoom: 14,
        maxZoom: 18,
    });
    loadBlocks();
    map.on('moveend', loadBlocks);
    console.log("Loaded");
}

$(document).ready(init);

/*
require([
    'json3', 'jquery', 'mapbox', 'mapbox.jquery', 'mapbox.jquery.geocoder',
    'mapbox.share', 'app/menu.jquery'
], init);
*/

