# proxy

## option 1

use input files/directories
-> requires a certain file naming strategy

<group>.<resource>.<name>.yaml

need to rework reconcilers and libraries -> to use the loaded input

## option 2

use database -> strategy for upstream right now
if we change this -> we take input and load this first + run reconciler
-> in a hierarchical upstream this means we need to load all upstreams hierachically first and load the deepest one, etc etc etc

## development environment

load as reconcilers/libraries files from the dev environment


TODO:
- load from dev environment (reconcilers and libraries)
- input loading -> there is no chnage -> only replace operation -> w/o metadata
    -> 
- rework upstream loading with the hierarchical approach


directories:
- crds: remain seperate -> to be able to load apis, but apis should not dynamically load
- libs and reconcilers:
    - they need to be preprocessed -> and put as input ()
- data:
    - 
- db: computed data
- 

## user

step1: choreoctl dev parse -> unifies reconciler/libraries in input directory
step2: choreoctl server start <path> or choreoctl server start + choreoctl server apply <ctx (url + ref + etc)>
step3: chorectl run once or run start/stop

## todo

hierarchical loading of upstreams