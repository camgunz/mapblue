var map = null;
var geoJSONLayer = null;
var selectedVotes = 0;

var blackCoeff = 0.4501479;
var hispanicCoeff = 0.077551;
var otherRaceCoeff = 0.1358834;
var unmarriedCoeff = 0.0911239;
var childlessCoeff = 0.115441;
var regressionConstant = 0.3638054;

var demColor = '#4488CC';
var repColor = '#BB4444';
var strokeColor = '#FFFF00';
var strokeWeight = 5;

var mapblueAPI = 'http://mapblue.org/lookup'
var openGeocoderAPI = 'http://nominatim.openstreetmap.org/search'
var censusGeocoderAPI =
    'http://geocoding.geo.census.gov/geocoder/locations/onelineaddress'
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

var totalVoters = 0;
var demVoters = 0;
var repVoters = 0;
var maxRepVoters = 0;
var maxDemVoters = 0;
var shadeOnVoteCounts = true;

function resetVoteTotals() {
    totalVoters = 0;
    demVoters = 0;
    repVoters = 0;
    maxRepVoters = 0;
    maxDemVoters = 0;
}

function calculateBlockStatistics(layer) {
    var block = layer.feature;

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
            (blackCoeff * block.properties.blackPct) +
            (hispanicCoeff * block.properties.hispanicPct) +
            (otherRaceCoeff * block.properties.otherRacePct) +
            (unmarriedCoeff * block.properties.unmarriedPct) +
            (childlessCoeff * block.properties.childlessPct);

    totalVoters += block.properties.over18;

    if (block.properties.demPct < .50) {
        block.properties.democrat = -(over18 * block.properties.demPct);
        repVoters += -(block.properties.democrat);
        if (-(block.properties.democrat) > maxRepVoters) {
            maxRepVoters = -(block.properties.democrat);
        }
    }
    else {
        block.properties.democrat = over18 * block.properties.demPct;
        demVoters += block.properties.democrat;
        if (block.properties.democrat > maxDemVoters) {
            maxDemVoters = block.properties.democrat;
        }
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
    var fillOpacity = 0.0;

    if (typeof block.properties.clicked == "undefined") {
        block.properties.clicked = false;
    }

    if (block.properties.over18 == 0) {
        return {
            fillColor: '#FFFFFF',
            fillOpacity: 0.0,
            stroke: block.properties.clicked,
            color: strokeColor,
            weight: strokeWeight
        };
    }

    if (block.properties.demPct < .5) {
        if (shadeOnVoteCounts) {
            fillOpacity = -(block.properties.democrat / maxRepVoters);
        }
        else {
            fillOpacity = 1.0 - block.properties.demPct;
        }

        if (fillOpacity < .15) {
            fillOpacity = .15;
        }

        return {
            fillColor: repColor,
            fillOpacity: fillOpacity,
            stroke: block.properties.clicked,
            color: strokeColor,
            weight: strokeWeight
        };
    }

    if (shadeOnVoteCounts) {
        fillOpacity = block.properties.democrat / maxDemVoters;
    }
    else {
        fillOpacity = block.properties.demPct;
    }

    if (fillOpacity < .15) {
        fillOpacity = .15;
    }

    return {
        fillColor: demColor,
        fillOpacity: fillOpacity,
        stroke: block.properties.clicked,
        color: strokeColor,
        weight: strokeWeight
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

function updateSelectedVotes() {
    $('#vote_counts').html(
        'Selected Votes: ' +
        Math.round(selectedVotes) + ' / ' +
        Math.round(demVoters) + ' (' +
        Math.round((selectedVotes / demVoters) * 100) + '%)'
    );
}

function loadBlocks() {
    $.getJSON(mapblueAPI, getMapCoordinates(), function(data) {
        geoJSONLayer.addData(data);
        resetVoteTotals();
        geoJSONLayer.eachLayer(calculateBlockStatistics);
        geoJSONLayer.setStyle(blockStyler);
        updateSelectedVotes();
    });
}

function reloadBlocks(e) {
    updateCoefficients();
    geoJSONLayer.eachLayer(calculateBlockStatistics);
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
        selectedVotes -= e.target.feature.properties.democrat;
    }
    else {
        e.target.feature.properties.clicked = true;
        selectedVotes += e.target.feature.properties.democrat;
    }

    e.target.setStyle({
        stroke: e.target.feature.properties.clicked
    });

    if (!L.Browser.ie && !L.Browser.opera) {
        e.target.bringToFront();
    }

    updateSelectedVotes();
}

function blockMousedOver(e) {
    var ps = e.target.feature.properties;

    $('#block_name').html(ps.name);
    $('#block_population').html(ps.over18);
    $('#block_black').html(
        Math.round(ps.blackPct * 100) +
        "% (" + Math.round(ps.black) + ")"
    );
    $('#block_hispanic').html(
        Math.round(ps.hispanicPct * 100) +
        "% (" + Math.round(ps.hispanic) + ")"
    );
    $('#block_other_race').html(
        Math.round(ps.otherRacePct * 100) +
        "% (" + Math.round(ps.otherRace) + ")"
    );
    $('#block_white').html(
        Math.round(ps.whitePct * 100) +
        "% (" + Math.round(ps.white) + ")"
    );
    $('#block_unmarried').html(
        Math.round(ps.unmarriedPct * 100) +
        "% (" + Math.round(ps.unmarried) + ")"
    );
    $('#block_childless').html(
        Math.round(ps.childlessPct * 100) +
        "% (" + Math.round(ps.childless) + ")"
    );
    $('#block_democratic').html(
        Math.round(ps.demPct * 100) +
        "% (" + Math.round(ps.democrat) + ")"
    );
}

function handleOpenGeocoderResponse(data) {
    $('#geocoder_loading').hide();
    if (data.length == 0) {
        $('#geocoder_error').html("Address not found");
        return;
    }

    var geolocation = data[0];

    if (geolocation.lat && geolocation.lon) {
        map.setView([geolocation.lat, geolocation.lon]);
    }
    else {
        $('#geocoder_error').html("Invalid geocoder response");
        $('#geocoder_error').show();
    }
}

function searchOpenForAddress() {
    $('#geocoder_error').hide();
    $('#geocoder_loading').show();
    $.getJSON(openGeocoderAPI, {
        format: 'json',
        q: $('#geocoder_address').val()
    }, handleOpenGeocoderResponse);
}

function handleCensusGeocoderResponse(data) {
    $('#geocoder_loading').hide();
    if ((!data) || (!data.addressMatches) || data.addressMatches.length == 0) {
        $('#geocoder_error').html("Address not found");
        $('#geocoder_error').show();
        return;
    }

    var coordinates = data.addressMatches[0].coordinates;
    var lat = coordinates.y;
    var lon = coordinates.x;

    map.setView([lat, lon]);
}

function searchCensusForAddress() {
    $('#geocoder_error').hide();
    $('#geocoder_loading').show();
    $.getJSON(censusGeocoderAPI, {
        format: 'jsonp',
        benchmark: 'Public_AR_Current',
        address: $('#geocoder_address').val()
    }, handleCensusGeocoderResponse);
}

function init() {
    $('#blackCoeff').val(blackCoeff);
    $('#hispanicCoeff').val(hispanicCoeff);
    $('#otherRaceCoeff').val(otherRaceCoeff);
    $('#unmarriedCoeff').val(unmarriedCoeff);
    $('#childlessCoeff').val(childlessCoeff);
    $('#regressionConstant').val(regressionConstant);

    $('#geocoder_address').keypress(function(e) {
        if (e.which == 13) {
            searchOpenForAddress();
        }
    });
    $('#geocoder_submit').click(searchOpenForAddress);
    $('#geocoder_loading').hide();
    $('#geocoder_error').hide();

    $('#regression_knobs').hide();
    $('#regression_button').click(function(e) {
        $('#regression_knobs').animate({width: 'toggle'});
    });

    $('#votes_knobs').hide();
    $('#votes_button').click(function(e) {
        $('#votes_knobs').animate({width: 'toggle'});
    });

    $('.shade_type_input').click(function(e) {
        var radioButton = $(e.target);
        var shadeValue = radioButton.attr('value');

        if (shadeValue == 'vote_counts') {
            shadeOnVoteCounts = true;
        }
        else {
            shadeOnVoteCounts = false;
        }
        reloadBlocks();
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

