$( document ).ready(function() {
  $.material.init();
  
  $( "#submit" ).click(function() {
    var url = $('#url');
    var newurl = $('#newurl');
    if (isUrl(url.val())){
      var posturl = "shorten?url="+encodeURIComponent(url.val())+"&newurl="+newurl.val()
      $.getJSON(posturl, function(data) {
        if ($.isPlainObject(data.error)) {
          error(data.error.message)
        } else {
          $('#url-dialog-url').text(url.val());
          $('#url-dialog-newurl').val($('#baseurl').text() + newurl.val());
          $('#url-dialog').modal('show');
          url.val('');
          newurl.val('');
        }
      });
    } else {
      error("Invalid url.");
    }
  });
  
  $( "#url-dialog-newurl" ).click(function() {
    document.getElementById("url-dialog-newurl").focus();
    document.getElementById("url-dialog-newurl").select();
  });
  
  $( "#rand" ).click(function() {
    var key = randKey();
    $('#newurl').focus().val(key);
  });
  
  function isUrl(s) {
    var regexp = /(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?/
    return regexp.test(s);
  }
  
  function randKey() {
    var n = 5
    var text = '';
    var possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';

    for(var i=0; i < n; i++)
    {
      text += possible.charAt(Math.floor(Math.random() * possible.length));
    }

    return text;
  }
  
});

function error(message){
  $('#error-dialog-message').text(message);
  $('#error-dialog').modal('show');
}
