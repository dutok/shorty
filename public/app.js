$( document ).ready(function() {
  $.material.init();
  loadRoute();
  
  $( "#submit" ).click(function() {
    var url = $('#url');
    var newurl = $('#newurl');
    var link = $('#baseurl').text() + newurl.val();
    if (isUrl(url.val())){
      var posturl = "shorten?url="+encodeURIComponent(url.val())+"&newurl="+newurl.val()
      $.getJSON(posturl, function(data) {
        if ($.isPlainObject(data.error)) {
          error(data.error.message)
        } else {
          $('#url-dialog-url').text(url.val());
          $('#url-dialog-newurl').val(link);
          $('#url-dialog').modal('show');
          url.val('');
          newurl.val('');
        }
      });
    } else {
      error("Invalid url.");
    }
  });
  
  $( "#analytics" ).click(function() {
    $('#analytics-dialog').modal('show');
  });
  
  $( "#analytics-get" ).click(function() {
    var alias = $("#analytics-input-alias");
    $.getJSON("u/"+alias.val()+"/analytics", function(data) {
      if ($.isPlainObject(data.error)) {
        error(data.error.message)
        $('#analytics-dialog').modal('hide');
      } else {
        console.log(data);
        $('#analytics-dialog-alias').text(alias.val());
        $('#analytics-dialog-url').text(data.url.url);
        $('#analytics-dialog-count').text(data.url.count);
        $('#analytics-panel').removeClass("hidden");
        alias.val('');
      }
    });
    $('#analytics-dialog').modal('show');
  });
  
  $( "#url-dialog-newurl" ).click(function() {
    document.getElementById("url-dialog-newurl").focus();
    document.getElementById("url-dialog-newurl").select();
  });
  
  $( "#rand" ).click(function() {
    var key = randKey();
    $('#newurl').focus().val(key);
  });
  
});

function loadRoute(){
  if(window.location.href.indexOf("404") > -1) {
    notFound();
  }
}

function error(message){
  $('#error-dialog-message').text(message);
  $('#error-dialog').modal('show');
}

function notFound(){
  $('#notfound-dialog').modal('show');
}

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