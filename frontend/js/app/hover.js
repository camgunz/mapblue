            // Make a grid layer and an empty GeoJSON layer
            var grid = L.mapbox.gridLayer('rclark.h2jc29lj');
            var geojson = L.geoJson(null);

            // Add those layers to the map
            geojson.addTo(map);
            grid.addTo(map);

            // When your mouse is over the gridlayer, read the GeoJSON that you put into your dataset 
            grid.on('mouseover', function (evt) {
                if (evt.data) {
                    var geo = JSON.parse(evt.data.geo);
                    geojson.clearLayers();
                    geojson.addData(geo);
                }
            });
            
            //https://a.tiles.mapbox.com/v3/rclark.h2jc29lj/page.html#9/36.1833/-110.3851
            //http://bl.ocks.org/rclark/7043524
            //https://www.mapbox.com/mapbox.js/example/v1.0.0/gridlayer-gridcontrol/