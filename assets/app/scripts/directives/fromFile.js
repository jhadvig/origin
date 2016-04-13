'use strict';

angular.module("openshiftConsole")
  .directive("fromFile", function($q,
                                  $uibModal,
                                  $location,
                                  $filter,
                                  TemplateService,
                                  TaskList,
                                  DataService,
                                  APIService) {
    return {
      restrict: "E",
      scope: false,
      templateUrl: "views/directives/from-file.html",
      controller: function($scope) {
        var aceEditorSession;
        var humanize = $filter('humanize');
        TaskList.clear();

        $scope.aceLoaded = function(editor) {
          aceEditorSession = editor.getSession();
          aceEditorSession.setOption('tabSize', 2);
          aceEditorSession.setOption('useSoftTabs', true);
          editor.setDragDelay = 0;
          editor.$blockScrolling = Infinity;

          $('.from-file .editor').animate({
            height: Math.floor(window.innerHeight * 0.50) + 'px'
          }, 30);
        };

        var aceDebounce = _.debounce(function(){
          try {
            JSON.parse($scope.editor.content);
            aceEditorSession.setMode("ace/mode/json");
          } catch (e) {
            try {
              YAML.parse($scope.editor.content);
              aceEditorSession.setMode("ace/mode/yaml");
            } catch (e) {
              return;
            }
          }
        }, 500);

        // Check if the editor isn't empty to disable the 'Add' button. Also check in what format the input is in (JSON/YAML) and change
        // the editor accordingly.
        $scope.aceChanged = function() {
          aceDebounce();
        };

        $scope.create = function() {
          delete $scope.alerts['create'];
          delete $scope.alerts['parsing'];
          var resource;

          // Trying to auto-detect what format the input is in. Since parsing JSON throws only SyntexError
          // exception if the string to parse is not valid JSON, it is tried first and then the YAML parser
          // is trying to parse the string. If that fails it will print the reason. In case the real reason
          // is JSON related the printed reason will be "Reason: Unable to parse", in case of YAML related
          // reason the true reason will be printed, since YAML parser throws an error object with needed
          // data.
          try {
            resource = JSON.parse($scope.editor.content);
          } catch (e) {
            try {
              resource = YAML.parse($scope.editor.content);
            } catch (e) {
              $scope.error = e;
              return;
            }
          }

          if (!resource.metadata || !resource.kind) {
            missingFieldsError(resource);
            return;
          }

          $scope.resourceKind = resource.kind;

          if ($scope.resourceKind === "Template") {
            $scope.templateOptions = {
              process: true,
              add: true
            };
          }

          if ($scope.resourceKind.endsWith("List")) {
            $scope.isList = true;
          }
          
          if ($scope.isList) {
            $scope.resourceList = resource.items;
            $scope.resourceName = '';
          } else {
            $scope.resourceList = [resource];
            $scope.resourceName = resource.metadata.name;
          }

          $scope.updateResources = [];
          $scope.createResources = [];

          var resourceCheckPromises = [];
          var errorOccured = false;
          _.forEach($scope.resourceList, function(item) {
            if (!item.metadata || !item.kind) {
              missingFieldsError(item);
              errorOccured = true;
              return false;
            } else if (item.metadata.namespace && item.metadata.namespace !== $scope.projectName) {
              $scope.error = {
                message: item.kind + " " + item.metadata.name + " can't be created in project " + item.metadata.namespace + ". Can't create resource in different projects."
              };
              errorOccured = true;
              return false;
            }
            resourceCheckPromises.push(checkResource(item));
          });

          if (!errorOccured) {
            $q.all(resourceCheckPromises).then(function() {
              // If resource if Template and it doesn't exist in the project
              if ($scope.createResources.length === 1 && $scope.resourceList[0].kind === "Template") {
                openTemplateProcessModal();
              // Else if any resources already exist
              } else if (!_.isEmpty($scope.updateResources)) {
                confirmReplace();
              } else {
                createAndUpdate();
              }
            });
          }
        };

        function missingFieldsError(item) {
          var missingField = "";
          if (!item.kind) {
            missingField = "'kind'";
          } else if (!item.metadata) {
            missingField = "'metadata'";
          }
          $scope.error = {
            message: "Resource is missing " + missingField + " field."
          };
        }

        function openTemplateProcessModal() {
          var modalInstance = $uibModal.open({
            animation: true,
            templateUrl: 'views/modals/process-template.html',
            controller: 'ProcessTemplateModalController',
            scope: $scope
          });
          modalInstance.result.then(function() {
            if ($scope.templateOptions.add) {
              createAndUpdate();
            } else {
              TemplateService.setTemplate($scope.resourceList[0]);
              redirect();
            }
          });      
        }

        function confirmReplace() {
          var modalInstance = $uibModal.open({
            animation: true,
            templateUrl: 'views/modals/confirm-replace.html',
            controller: 'ConfirmReplaceModalController',
            scope: $scope
          });
          modalInstance.result.then(function() {
            createAndUpdate();
          });
        }

        // create 
        function createAndUpdate() {
          var createUpdatePromises = [];
          if ($scope.updateResources.length > 0) {
            createUpdatePromises.push(updateResourceList());
          }
          if ($scope.createResources.length > 0) {
            createUpdatePromises.push(createResourceList());
          }
          $q.all(createUpdatePromises).then(function() {
            redirect();
          });
        }

        function redirect() {
          var subPath;
          if ($scope.resourceKind === "Template" && $scope.templateOptions.process) {
            subPath =  "create/fromtemplate?name=" + $scope.resourceName + "&namespace=" + encodeURIComponent($scope.projectName);
          } else {
            subPath = "overview";
          }
          $location.url("project/" + encodeURIComponent($scope.projectName) + "/" + subPath);
          // $scope.$evalAsync();
        }

        function checkResource(item) {
          // Check if the resource already exists. If it does, replace it spec with the new one.
          return DataService.get(APIService.kindToResource(item.kind), item.metadata.name, $scope.context, {errorNotification: false}).then(
            // resource does exist
            function(resource) {
              resource.spec = item.spec;
              $scope.updateResources.push(resource);
            },
            // resource doesn't exist
            function() {
              $scope.createResources.push(item);
          });
        }

        function createResourceList(){
          var titles = {
            started: "Creating in project " + $scope.projectName,
            success: "Created in project " + $scope.projectName,
            failure: "Failed to create in project " + $scope.projectName
          };
          var helpLinks = {};
          TaskList.add(titles, helpLinks, function() {
            var d = $q.defer();

            DataService.listAction($scope.createResources, $scope.context).then(
              function(result) {
                var alerts = [];
                var hasErrors = false;
                if (result.failure.length > 0) {
                  hasErrors = true;
                  result.failure.forEach(
                    function(failure) {
                      alerts.push({
                        type: "error",
                        message: "Cannot create " + humanize(failure.object.kind).toLowerCase() + " \"" + failure.object.metadata.name + "\". ",
                        details: failure.data.message
                      });
                    }
                  );
                  result.success.forEach(
                    function(success) {
                      alerts.push({
                        type: "success",
                        message: "Created " + humanize(success.kind).toLowerCase() + " \"" + success.metadata.name + "\" successfully. "
                      });
                    }
                  );
                } else {
                  var alertMsg;
                  if ($scope.isList) {
                    alertMsg = "All items in list were created successfully.";
                  } else {
                    alertMsg = $scope.resourceKind + " " + $scope.resourceName + " was successfully created.";
                  }
                  alerts.push({ type: "success", message: alertMsg});
                }
                d.resolve({alerts: alerts, hasErrors: hasErrors});
              }
            );
            return d.promise;
          });
        }


        function updateResourceList(){
          var titles = {
            started: "Updating in project " + $scope.projectName,
            success: "Updated in project " + $scope.projectName,
            failure: "Failed to update in project " + $scope.projectName
          };
          var helpLinks = {};
          TaskList.add(titles, helpLinks, function() {
            var d = $q.defer();

            DataService.listAction($scope.updateResources, $scope.context, {action: "update"}).then(
              function(result) {
                var alerts = [];
                var hasErrors = false;
                if (result.failure.length > 0) {
                  hasErrors = true;
                  result.failure.forEach(
                    function(failure) {
                      alerts.push({
                        type: "error",
                        message: "Cannot update " + humanize(failure.object.kind).toLowerCase() + " \"" + failure.object.metadata.name + "\". ",
                        details: failure.data.message
                      });
                    }
                  );
                  result.success.forEach(
                    function(success) {
                      alerts.push({
                        type: "success",
                        message: "Updated " + humanize(success.kind).toLowerCase() + " \"" + success.metadata.name + "\" successfully. "
                      });
                    }
                  );
                } else {
                  var alertMsg;
                  if ($scope.isList) {
                    alertMsg = "All items in list were updated successfully.";
                  } else {
                    alertMsg = $scope.resourceKind + " " + $scope.resourceName + " was successfully updated.";
                  }
                  alerts.push({ type: "success", message: alertMsg});
                }
                d.resolve({alerts: alerts, hasErrors: hasErrors});
              }
            );
            return d.promise;
          });
        }
      }
    };
  });