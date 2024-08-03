# evrcat

`evrcat` finds all hex strings of known EchoVR symbols into their string representation. It uses the symbol cache of <https://github.com/echotools/nakama>.

## Usage

## Example

It translates this:

```log
[NETLOBBY] Starting session (gametype 0x33BBF6842DF97A3F, level 0xEA49CDE332AA536F)
Dedicated: beginning session
[NETGAME] Session Started
[NETGAME] Loading map '0xEA49CDE332AA536F'
[NETGAME] NetGame switching state (from lobby, to loading level)
[LEVELLOAD] Loading level '0x4D82118C7C91B6BB'
[NETGAME] LobbyOwnerDataChangedCB
[LEVELLOAD] Level '0x4D82118C7C91B6BB' offset by (0.000, 0.000, 0.000)
[BLACKBOARD] Unable to resolve script actor for blackboard value 0xF565D0591C4B5CDB on actor 0x8FC4E87041FD4411
[LEVELLOAD] Finished loading level '0x4D82118C7C91B6BB' in 40 ms
```

into this:

```log
[NETLOBBY] Starting session (gametype echo_combat_private, level pty_mpl_combat_pebbles)
Dedicated: beginning session
[NETGAME] Session Started
[NETGAME] Loading map 'pty_mpl_combat_pebbles'
[NETGAME] NetGame switching state (from lobby, to loading level)
[LEVELLOAD] Loading level 'mnu_master_mp_ingame'
[NETGAME] LobbyOwnerDataChangedCB
[LEVELLOAD] Level 'mnu_master_mp_ingame' offset by (0.000, 0.000, 0.000)
[BLACKBOARD] Unable to resolve script actor for blackboard value 0xf565d0591c4b5cdb on actor 0x8fc4e87041fd4411
[LEVELLOAD] Finished loading level 'mnu_master_mp_ingame' in 40 ms
```
