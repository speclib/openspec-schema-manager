# Keys

Generated from the code by `go test ./internal/tui -run TestWriteKeyReference -update`.
Every screen reports its own keys, and the help overlay inside ossm reads the same list.

## Everywhere

| Key             | Does                         |
| --------------- | ---------------------------- |
| tab / shift+tab | move between tabs            |
| 1 to 4          | select a tab by number       |
| : or ctrl+o     | open a schema folder by path |
| ?               | open and close this help     |
| q               | close an overlay, or quit    |
| ctrl+c          | quit from anywhere           |

## Project

| Key       | Does                                           |
| --------- | ---------------------------------------------- |
| up / down | move through the schemas                       |
| enter     | open the selected schema                       |
| s         | set the selected schema as the project default |
| u         | compare with the registry                      |
| p         | send to the composer                           |
| r         | read the project again                         |

## Registry

| Key       | Does                                 |
| --------- | ------------------------------------ |
| enter     | open the selected schema             |
| i         | install into this project            |
| p         | send to the composer                 |
| up / down | move the selection                   |
| /         | filter over id, name and description |
| esc       | clear the filter                     |
| r         | refresh the registry now             |

## Local

| Key       | Does                       |
| --------- | -------------------------- |
| up / down | move through the schemas   |
| enter     | open the selected schema   |
| t         | browse the schema's files  |
| c         | duplicate under a new name |
| p         | send to the composer       |
| r         | scan the directories again |

## Composer

| Key          | Does                                       |
| ------------ | ------------------------------------------ |
| tab is taken | use left / right to move between the panes |
| a            | add the palette selection to the canvas    |
| x            | remove the selected artifact               |
| l            | link the selected artifact to another      |
| u            | remove a link                              |
| g            | toggle the apply gate                      |
| t            | set the tracked file                       |
| R            | rename the selected artifact               |
| d            | show and hide the diagram                  |
| w            | write the schema                           |
| s            | save a draft                               |
| o            | open a draft                               |

## Modes

Some screens change what the keys mean while they are taking text or offering a choice.
The help overlay follows, so `?` always lists what works where you are standing.

### Local, browsing a schema's files

| Key       | Does                    |
| --------- | ----------------------- |
| up / down | move through the files  |
| e         | edit the selected file  |
| t         | back to the schema list |

### Local, naming a duplicate

| Key   | Does           |
| ----- | -------------- |
| enter | write the copy |
| esc   | cancel         |

### Composer, choosing what to link to

| Key       | Does                   |
| --------- | ---------------------- |
| up / down | choose what to require |
| enter     | add the link           |
| esc       | cancel                 |

### Composer, choosing a draft

| Key       | Does           |
| --------- | -------------- |
| up / down | choose a draft |
| enter     | resume it      |
| esc       | cancel         |

### Composer, any prompt

| Key   | Does    |
| ----- | ------- |
| enter | confirm |
| esc   | cancel  |

### Registry, confirming an install

| Key     | Does                            |
| ------- | ------------------------------- |
| y       | confirm the install             |
| o       | overwrite what is already there |
| n / esc | cancel                          |

### Anywhere, the path prompt

| Key   | Does              |
| ----- | ----------------- |
| enter | open the folder   |
| tab   | complete the path |
| esc   | cancel            |

