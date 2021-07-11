// Captive portal Javascript
// by Stefan Midjich @ Cygate AB
//

var debug = true;

function getUrlParameter(sParam, default_value) {
    var sPageURL = decodeURIComponent(window.location.search.substring(1)),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : sParameterName[1];
        }
    }

    return default_value;
}


// This function ensures the user gets redirect to the correct destination once
// all jobs have succeeded in the portal software.
function do_success() {
    var url = getUrlParameter('url', 'www.google.com');

    // If url does not start with http the window.location redirect
    // won't work. So prefix http to url.
    if (!url.startsWith('http')) {
        url = 'http://' + url;
    }
    //console.log('success: ' + url);
    $('#error-box').html('<p>If you\'re not automatically redirected <a href="https://www.google.com/">click here</a>.</p>');
    $('#error-box').show();
    $('#statusDiv').html('');
    $('#approveButton').prop('disabled', false);

    // Redirect user to the url paramter.
    window.location = url;
}


// Show an error to the user
function do_error(message) {
    $('#approveButton').prop('disabled', false);
    $('#statusDiv').html('');

    $('#error-box').show();
    $('#error-box').html('<p>Failed. Reload page and try again or contact support.</p> ');
    if (message) {
        //console.log(message);
    }
}
