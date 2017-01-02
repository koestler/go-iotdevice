$(document).ready(function() {

    var pollApi = function() {
        console.log('poll begin');

        $.get("device/1", function( data ) {
            console.log(data);
        });
    }
    setInterval(pollApi, 1000);

});
