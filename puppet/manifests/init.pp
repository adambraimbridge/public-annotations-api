class public_annotations_api {

  $configParameters = hiera('configParameters','')

  class { "go_service_profile" :
    service_module => $module_name,
    service_name => 'public-annotations-api',
    configParameters => $configParameters
  }

}
