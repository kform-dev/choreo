# reconciler design

## init

input
- name
- apigroups
- client
- informerFactory
- we need a file

initialize:
- create queue
- call register to know what the reconciler is about 
    - validate (check if apis exist, etc)
    - add eventHandler to informerfactory

register:
1. get reconciler data
    get the data:
        - starlark: initiate the fn
        - kform: get the register file
        - maybe we have to run a function()
    parse/validate the data
    options:
    1. provide dedicated input in the reconciler
        - independent from the code
        + generic: works for any language consuming it
    2. call a specific method to retrieve the data.

    Do we do this during validation phase:
    + in the geenric way this is nice
    - if we need to instantiate code/operator this is cumbersome 
2. register the eventHandlers with this data

## execContext

for resource
- update: can add metadata, normally no spec update
- updateStatus: status only
own resource
- create: unstructured object
- apply: no parameter
- delete: no parameter
any resource
- get gvk, nsn
- list gvk fieldSelector