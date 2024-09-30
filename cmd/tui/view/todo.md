



## selecting a new branch

we assume branches
0: main
1: checkout branch

how to select other branches?
-> branch page

we setup a branch streamer to new input -> what do we do with it?
-> if setup for checkout: we follow checkout
    - there should always be a checkout (main or other)
    - if checkout changes we need to reset views
-> if setup for main: we continue to follow main (no changes)
-> if we setup for other, what do we do if it gets deleted
    - reset pages and move to menu


## approaches:

1. use a global ticker
    - update page components
    - which pages are active
    - you need to manage focus

2. use a local ticker for the page that is active
    - does not allow to update branches while

## if menu is selected can we select global focus