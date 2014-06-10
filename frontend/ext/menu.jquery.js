jQuery(document).ready(function() {
  jQuery(".content").hide();
  //toggle the componenet with class msg_body
  jQuery(".hide").click(function()
  {
    jQuery(this).next(".content").slideToggle(10);
  });
});