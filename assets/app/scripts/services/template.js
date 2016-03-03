'use strict';

angular.module("openshiftConsole")
  .service("TemplateService", function(){

    var template = {};
    return {
      setTemplate: function(temp) {
        template = temp;
      },
      getTemplate: function() {
        return template;
      },
      clearTemplate: function() {
        template = {};
      }
    };

  });