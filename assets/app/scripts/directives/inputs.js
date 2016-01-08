'use strict';

angular.module("openshiftConsole")
  .directive("sourceInput", function () {
    return {
    	restrict: "E",
    	scope: {
    		model: '=',
    		comparedType: "=",
    		label: "@"
    	},
    	templateUrl: "views/directives/source-input.html",
    	replace: true,
    	link: function() {
    		var sourceURLPattern = /^((ftp|http|https|git):\/\/(\w+:{0,1}[^\s@]*@)|git@)?([^\s@]+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?$/;
    	}
    }
});