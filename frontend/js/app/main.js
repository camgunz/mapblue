var map = null;
var geoJSONLayer = null;

var blackCoeff = 0.4501479;
var hispanicCoeff = 0.077551;
var otherRaceCoeff = 0.1358834;
var unmarriedCoeff = 0.0911239;
var childlessCoeff = 0.115441;
var regressionConstant = 0.3638054;

var mapblueAPI = 'http://mapblue.org/lookup'
var geocoderAPI = 'http://nominatim.openstreetmap.org/search'
var stateHouseLatitude = 39.768732;
var stateHouseLongitude = -86.162612;
// var tileJSON = 'https://a.tiles.mapbox.com/v3/examples.map-20v6611k,mapbox.dc-property-values.jsonp?secure';
var tileJSON = 'http://{s}.tile.stamen.com/terrain/{z}/{x}/{y}.png'
var tileJSONAttribution = 
    'Map tiles by <a href="http://stamen.com">Stamen Design</a>, ' +
    '<a href="http://creativecommons.org/licenses/by/3.0">CC BY 3.0</a> ' +
    '&mdash; Map data &copy; ' +
    '<a href="http://openstreetmap.org">OpenStreetMap</a> contributors, ' +
    '<a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>'
var blockIDs = [];

function calculateBlockStatistics(block) {
    var over18 = block.properties.over18;
    var black = block.properties.black;
    var hispanic = block.properties.hispanic;
    var otherRace = block.properties.otherRace;
    var nonWhite = black + hispanic + otherRace;
    var white = over18 - nonWhite;
    var unmarried = block.properties.unmarried;
    var childless = block.properties.childless;

    if (over18 == 0) {
        block.properties.white = 0;
        block.properties.blackPct = 0;
        block.properties.hispanicPct = 0;
        block.properties.otherRacePct = 0;
        block.properties.whitePct = 0;
        block.properties.unmarriedPct = 0;
        block.properties.childlessPct = 0;
        block.properties.demPct = 0;
        block.properties.democrat = 0;
        return;
    }

    block.properties.white = white;
    block.properties.blackPct = black / over18;
    block.properties.hispanicPct = hispanic / over18;
    block.properties.otherRacePct = otherRace / over18;
    block.properties.whitePct = white / over18;
    block.properties.unmarriedPct = unmarried / over18;
    block.properties.childlessPct = childless / over18;
    block.properties.demPct = regressionConstant +
        (blackCoeff     * block.properties.blackPct)     +
        (hispanicCoeff  * block.properties.hispanicPct)  +
        (otherRaceCoeff * block.properties.otherRacePct) +
        (unmarriedCoeff * block.properties.unmarriedPct) +
        (childlessCoeff * block.properties.childlessPct);
    if (block.properties.demPct < .50) {
        block.properties.democrat = -(over18 * block.properties.demPct);
    }
    else {
        block.properties.democrat = over18 * block.properties.demPct;
    }
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
    if (typeof block.properties.clicked == "undefined") {
        block.properties.clicked = false;
    }

    calculateBlockStatistics(block);

    if (block.properties.over18 == 0) {
        return {
            fillColor: '#FFFFFF',
            fillOpacity: 0.0,
            stroke: block.properties.clicked,
            weight: 2
        };
    }

    if (block.properties.demPct < .5) {
        return {
            fillColor: '#BB4444',
            fillOpacity: 1.0 - block.properties.demPct,
            stroke: block.properties.clicked,
            weight: 2
        };
    }

    return {
        fillColor: '#4488CC',
        fillOpacity: block.properties.demPct,
        stroke: block.properties.clicked,
        weight: 2
    };
}

function updateCoefficients() {
    blackCoeff = parseFloat($('#blackCoeff').val());
    hispanicCoeff = parseFloat($('#hispanicCoeff').val());
    otherRaceCoeff = parseFloat($('#otherRaceCoeff').val());
    unmarriedCoeff = parseFloat($('#unmarriedCoeff').val());
    childlessCoeff = parseFloat($('#childlessCoeff').val());
    regressionConstant = parseFloat($('#regressionConstant').val());
}

function loadBlocks() {
    $.getJSON(mapblueAPI, getMapCoordinates(), function(data) {
        geoJSONLayer.addData(data);
    });
}

function reloadBlocks(e) {
    updateCoefficients();
    geoJSONLayer.setStyle(blockStyler);
}

function blockFilter(block) {
    if (blockIDs.indexOf(block.id) == -1) {
        blockIDs.push(block.id);
        return true;
    }
    return false;
}

function blockClicked(e) {
    if (e.target.feature.properties.clicked) {
        e.target.feature.properties.clicked = false;
    }
    else {
        e.target.feature.properties.clicked = true;
    }
    e.target.setStyle({
        stroke: e.target.feature.properties.clicked
    });
    if (!L.Browser.ie && !L.Browser.opera) {
        layer.bringToFront();
    }
}

function blockMousedOver(e) {
    var ps = e.target.feature.properties;

    console.log(ps);

    $('#block_name').html(ps.name);
    $('#block_population').html(ps.over18);
    $('#block_black').html(
        Math.round(ps.blackPct * 100) + "% (" + Math.round(ps.black) + ")"
    );
    $('#block_hispanic').html(
        Math.round(ps.hispanicPct * 100) + "% (" + Math.round(ps.hispanic) + ")"
    );
    $('#block_other_race').html(
        Math.round(ps.otherRacePct * 100) + "% (" + Math.round(ps.otherRace) + ")"
    );
    $('#block_white').html(
        Math.round(ps.whitePct * 100) + "% (" + Math.round(ps.white) + ")"
    );
    $('#block_unmarried').html(
        Math.round(ps.unmarriedPct * 100) + "% (" + Math.round(ps.unmarried) + ")"
    );
    $('#block_childless').html(
        Math.round(ps.childlessPct * 100) + "% (" + Math.round(ps.childless) + ")"
    );
    $('#block_democratic').html(
        Math.round(ps.demPct * 100) + "% (" + Math.round(ps.democrat) + ")"
    );
}

function searchForAddress() {
    $.getJSON(
        geocoderAPI,
        {format: 'json', q: $('#geocoder_address').val()},
        function(data) {
            if (data.length == 0) {
                $('#geocoder_error').html("Address not found");
                return;
            }

            var geolocation = data[0];

            console.log(geolocation);
            if (geolocation.lat && geolocation.lon) {
                map.setView([geolocation.lat, geolocation.lon]);
            }
            else {
                $('#geocoder_error').html("Invalid geocoder response");
            }
        }
    );
}

function init() {
    $('#blackCoeff').val(blackCoeff);
    $('#hispanicCoeff').val(hispanicCoeff);
    $('#otherRaceCoeff').val(otherRaceCoeff);
    $('#unmarriedCoeff').val(unmarriedCoeff);
    $('#childlessCoeff').val(childlessCoeff);
    $('#regressionConstant').val(regressionConstant);

    $('#geocoder_address').keypress(function (e) {
        if (e.which == 13) {
            searchForAddress();
        }
    });
    $('#geocoder_submit').click(searchForAddress);
    $('#regression_knobs').hide();
    $('#regression_button').click(function(e) {
        $('#regression_knobs').animate({width: 'toggle'});
    });

    map = L.map('map').setView(
        [stateHouseLatitude, stateHouseLongitude], 16
    );
    map.on('moveend', loadBlocks);

    L.tileLayer(tileJSON, {
        attribution: tileJSONAttribution,
        subdomains: 'abcd',
        minZoom: 13,
        maxZoom: 18
    }).addTo(map);

    map.zoomControl.setPosition('topright');

    geoJSONLayer = L.geoJson(null, {
        style: blockStyler,
        filter: blockFilter,
        onEachFeature: function(block, layer) {
            layer.on({
                click: blockClicked,
                mouseover: blockMousedOver
            })
        }
    });
    geoJSONLayer.addTo(map);
    loadBlocks();
}

$(document).ready(init);

