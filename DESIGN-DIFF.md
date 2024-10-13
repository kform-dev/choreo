## Data Storage and Versioning:
- in memory snapshot

## Diff generation

## Diff representation

- human/machine readable

## UI/experiences



snapshotter

How can data be updated in the database ?
- only be the runner
- we can create an initial snapshot when the project gets loaded or after the first run
    -> we store the apis and input + database info; reference is per run if the data changes; so we store per run #, time and optional name reference and result
    -> 



TBD: How do we change api(s) ????
-> today they are only changed when we start the server, but maybe we need more flexibility


TODO:
- move snapshotter to its own grpc service
- delete fails
- why dont we show choreoAPI resources when building the inventory
