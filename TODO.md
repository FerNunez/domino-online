# TODO

## BACKEND

- Change the first connection/reconnection fetching game in the websocket. I don't like that when there is no game started in lobby, when users join lobby it gets error "no game".. Maybe we should switch to pull game state by frontend instead of autoamtically fetch? 
- Decide what to do for the frontend localStorage vs sessionStorage vs backend cookies


## FRONTEND

- Improve the Result pane (when a round finished)
  - Move the round result inside the board to. Like a pop up or similar
  - Improve the TEAM ID scores.. maybe add players name in the TeamID and add like highlits?

- Improve the round history/ action history:
  - Maybe add the starting hand as vertical tiles that are rodered/grouped  for most repeated if possible
  - make the scrollable part more compact in width.
  - maybe add the whole board in the action history where you hover a tile action and it is shown in the board to point out when was played (and maybe add like highlighting of board pieces that were played before that tile and those player for other tiles)
  - maybe that a highlight coloring to separated like tiles from team or from players?
