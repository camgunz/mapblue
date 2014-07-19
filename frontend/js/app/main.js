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
var demStrokeColor = '#224466';
var repColor = '#BB4444';
var repStrokeColor = '#552222';
var strokeWeight = 2;
var selectedStrokeColor = '#FFFF00';
var selectedStrokeWeight = 5;
var minOpacity = .15;

var mapblueAPI = 'http://mapblue.org/lookup'
var openGeocoderAPI = 'http://nominatim.openstreetmap.org/search'
var censusGeocoderAPI =
    'http://geocoding.geo.census.gov/geocoder/locations/onelineaddress'
var stateHouseLatitude = 39.768732;
var stateHouseLongitude = -86.162612;
var tileJSON = 'http://{s}.tile.stamen.com/terrain/{z}/{x}/{y}.png'
var tileJSONAttribution =
    'Map tiles by <a href="http://stamen.com">Stamen Design</a>, ' +
    '<a href="http://creativecommons.org/licenses/by/3.0">CC BY 3.0</a> ' +
    '&mdash; Map data &copy; ' +
    '<a href="http://openstreetmap.org">OpenStreetMap</a> contributors, ' +
    '<a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>'

var seenBlockIDs = [];
var lastCoords = null;

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
    selectedVotes = 0;
}

function addBlockStats(block) {
    totalVoters += block.properties.over18;
    repVoters += block.properties.repVotes;
    demVoters += block.properties.demVotes;

    if (block.properties.netVotes < 0) {
        if ((-block.properties.netVotes) > maxRepVoters) {
            maxRepVoters = -block.properties.netVotes;
        }
    }
    else {
        if (block.properties.netVotes > maxDemVoters) {
            maxDemVoters = block.properties.netVotes;
        }
    }
}

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
        block.properties.demVotes = 0;
        block.properties.repPct = 0;
        block.properties.repVotes = 0;
        block.properties.netVotes = 0;
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
    block.properties.repPct = 1.0 - block.properties.demPct;
    block.properties.demVotes = over18 * block.properties.demPct;
    block.properties.repVotes = over18 - block.properties.demVotes;
    block.properties.netVotes =
        block.properties.demVotes - block.properties.repVotes;
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
    var fillColor = '#000000';
    var fillOpacity = 0.1;
    var stroke = false;
    var color = '#FFFFFF';
    var weight = 0;

    if (block.properties.clicked) {
        color = selectedStrokeColor,
        weight = selectedStrokeWeight
    }

    if (block.properties.over18 > 0) {
        if (block.properties.demPct < .5) {
            if (shadeOnVoteCounts) {
                fillOpacity = -(block.properties.repVotes / maxRepVoters);
            }
            else {
                fillOpacity = 1.0 - block.properties.demPct;
            }

            fillColor = repColor;
            stroke = true;
            if (!block.properties.clicked) {
                color = repStrokeColor;
                weight = strokeWeight;
            }
        }
        else {
            if (shadeOnVoteCounts) {
                fillOpacity = block.properties.netVotes / maxDemVoters;
            }
            else {
                fillOpacity = block.properties.demPct;
            }

            fillColor = demColor;
            stroke = true;
            if (!block.properties.clicked) {
                color = demStrokeColor;
                weight = strokeWeight;
            }
        }

        if (fillOpacity < minOpacity) {
            fillOpacity = minOpacity;
        }
    }

    return {
        fillColor: fillColor,
        fillOpacity: fillOpacity,
        stroke: stroke,
        color: color,
        weight: weight
    };
}

function updateCoefficients() {
    blackCoeff = parseFloat($('#black_coeff').val());
    hispanicCoeff = parseFloat($('#hispanic_coeff').val());
    otherRaceCoeff = parseFloat($('#other_race_coeff').val());
    unmarriedCoeff = parseFloat($('#unmarried_coeff').val());
    childlessCoeff = parseFloat($('#childless_coeff').val());
    regressionConstant = parseFloat($('#regression_constant').val());
}

function updateSelectedVotes() {
    $('#total_votes').html(Math.round(totalVoters));
    $('#dem_votes').html(Math.round(demVoters));
    $('#rep_votes').html(Math.round(repVoters));
    $('#selected_votes').html(Math.round(selectedVotes));
}

function loadBlocks() {
    var firstTime = lastCoords == null;
    var coords = getMapCoordinates();

    if ((!firstTime) &&
        coords.lat1 == lastCoords.lat1 &&
        coords.lon1 == lastCoords.lon1 &&
        coords.lat2 == lastCoords.lat2 &&
        coords.lon2 == lastCoords.lon2) {
        return;
    }

    lastCoords = coords;

    if (!firstTime) {
        $('#title').text('');
        $('#title').toggleClass('loading');
    }

    $.getJSON(mapblueAPI, coords, function(data) {
        var features = [];
        var voters = 0;
        var loadedBlockIDs = [];

        resetVoteTotals();

        for (var i = 0; i < data.features.length; i++) {
            var feature = data.features[i];

            calculateBlockStatistics(feature);
            addBlockStats(feature);
            voters += feature.properties.over18;
            loadedBlockIDs.push(feature.id);

            if (seenBlockIDs.indexOf(feature.id) == -1) {
                seenBlockIDs.push(feature.id);
                features.push(feature);
            }
        }

        data.features = features;

        geoJSONLayer.addData(data);
        geoJSONLayer.eachLayer(function(layer) {
            var block = layer.feature;

            if (typeof block.properties.clicked == "undefined") {
                block.properties.clicked = false;
            }
            else if (loadedBlockIDs.indexOf(block.id) == -1) {
                if (block.properties.clicked) {
                    block.properties.clicked = false;
                }
            }
            else if (block.properties.clicked) {
                selectedVotes += block.properties.netVotes;
            }
        });
        geoJSONLayer.setStyle(blockStyler);
        updateSelectedVotes();

        if (!firstTime) {
            $('#title').toggleClass('loading');
            $('#title').text('Map Blue');
        }
    });

}

function reloadBlocks(e) {
    updateCoefficients();
    geoJSONLayer.eachLayer(function(layer) {
        calculateBlockStatistics(layer.feature);
    });
    geoJSONLayer.setStyle(blockStyler);
}

function blockClicked(e) {
    if (e.target.feature.properties.clicked) {
        selectedVotes -= e.target.feature.properties.netVotes;
    }
    else {
        selectedVotes += e.target.feature.properties.netVotes;
    }

    e.target.feature.properties.clicked = !e.target.feature.properties.clicked;
    geoJSONLayer.setStyle(blockStyler);

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
        "% (" + Math.round(ps.demVotes) + ")"
    );
    $('#block_republican').html(
        Math.round(ps.repPct * 100) +
        "% (" + Math.round(ps.repVotes) + ")"
    );
    $('#block_net').html(Math.round(ps.netVotes));
}

function handleOpenGeocoderResponse(data) {
    $('#geocoder_submit').removeClass('loading').addClass('search');

    if (data.length == 0) {
        $('#geocoder_error').html("Address not found");
        $('#geocoder_error').show();
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
    $('#geocoder_submit').removeClass('search').addClass('loading');
    $.getJSON(openGeocoderAPI, {
        format: 'json',
        q: $('#geocoder_address').val()
    }, handleOpenGeocoderResponse);
}

function handleCensusGeocoderResponse(data) {
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
    $.getJSON(censusGeocoderAPI, {
        format: 'jsonp',
        benchmark: 'Public_AR_Current',
        address: $('#geocoder_address').val()
    }, handleCensusGeocoderResponse);
}

function init() {
    $('#black_coeff').val(blackCoeff);
    $('#hispanic_coeff').val(hispanicCoeff);
    $('#other_race_coeff').val(otherRaceCoeff);
    $('#unmarried_coeff').val(unmarriedCoeff);
    $('#childless_coeff').val(childlessCoeff);
    $('#regression_constant').val(regressionConstant);

    map = L.map('map').setView(
        [stateHouseLatitude, stateHouseLongitude], 16
    );
    map.on('moveend', loadBlocks);

    L.tileLayer(tileJSON, {
        attribution: tileJSONAttribution,
        subdomains: 'abcd',
        minZoom: 12,
        maxZoom: 18
    }).addTo(map);

    map.zoomControl.setPosition('topright');
    map.attributionControl.setPosition('bottomright');

    $('#geocoder_address').keypress(function(e) {
        if (e.which == 13) {
            searchOpenForAddress();
        }
    });
    $('#geocoder_submit').click(searchOpenForAddress);

    $('#about_dialog').dialog({
        autoOpen: false,
        modal: true,
        width: 400,
        height: 300
    });
    $('#about').click(function (e) { $('#about_dialog').dialog('open'); });

    $('#config_dialog').dialog({
        autoOpen: false,
        modal: true,
        width: 400
    });
    $('#config').click(function (e) { $('#config_dialog').dialog('open'); });

    $('.shade_type_input').click(function (e) {
        shadeOnVoteCounts = ($(e.target).attr('value') == 'vote_counts');
        reloadBlocks();
    });

    geoJSONLayer = L.geoJson(null, {
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

