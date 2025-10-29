local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'orderservice',
];

local proto = [
    'api/client/testinternal/testinternal.proto',
    'api/server/orderserviceinternal/orderserviceinternal.proto',
];

project.project(appIDs, proto)