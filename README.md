TwitchChatArchiver ![MrDestructoid][0]
===

This is the very simple "chat bot" I use for archiving messages during events.

## Important remarks
* Every channel gets its own connection. In a normal situation this would be
    completely pointless, but Twitch will drop a connection if too many
    messages are in the sending queued to be sent to it and the high volume
    nature of rooms during events means that limit is easily hit.
* The exact maximum number of connections allowed per IP address isn't
    officially documented, but as an anti-spam measure Twitch will drop all of
    your connections and possibly further temporarily ban your IP should you
    open too many. If you want to archive a lot of channels (≥ 50) then
    this project probably isn't for you, but you presumably know that.

## Configuration

* `nick`: The username of the account you want to log in with.
* `oauth`: An oauth token -- [get one here][1]
* `channels`: The channels you want the join. Note that they **must** be
    prefixed with '#.'

The directory structure generated is as follows:

    logs/
        faceittv/
            1536105600.txt
            1536105600_userstate.txt
        starladder5/
            1536192000.txt
            1536192000_userstate.txt
        99damage/
            1536278400.txt
            1536278400_userstate.txt

Note that the filenames are UTC Unix timestamps.

[0]: https://static-cdn.jtvnw.net/emoticons/v1/28/1.0
[1]: https://twitchapps.com/tmi/
[2]: https://dev.twitch.tv/docs/irc/tags/#privmsg-twitch-tags
[3]: https://dev.twitch.tv/docs/irc/tags/#usernotice-twitch-tags
