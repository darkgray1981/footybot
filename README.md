# footybot
Football information dispensing IRC bot

For a long time, this bot has been serving #epl with pluck, but BBC went and changed their APIs, so nothing works anymore. It's up to you (or me, or someone) to either puzzle out a way to parse the new API, or to find something better, in order to resurrect its full functionality.

While the old BBC data could be simply parsed from the web tables, the new one uses React in combination with an API server, returning blobs of JSON.

The general scores and fixturs page is located at  
http://www.bbc.com/sport/football/scores-fixtures  
which makes queries in the form of:  
http://push.api.bbci.co.uk/p?t=morph%3A%2F%2Fdata%2Fbbc-morph-football-scores-match-list-data%2FendDate%2F2017-07-05%2FstartDate%2F2017-07-05%2Ftournament%2Ffull-priority-order%2Fversion%2F2.2.1%2FwithPlayerActions%2Ffalse

There are then individual team pages located at  
http://www.bbc.com/sport/football/teams/arsenal/scores-fixtures  
which make queries in the form of:  
http://push.api.bbci.co.uk/p?t=morph%3A%2F%2Fdata%2Fbbc-morph-football-scores-match-list-data%2FendDate%2F2018-08-31%2FstartDate%2F2017-08-01%2Fteam%2Farsenal%2Fversion%2F2.2.1%2FwithPlayerActions%2Ffalse

A formatted sample of the returned data: https://hastebin.com/wilejivewi.json


The league tables section has changed as well, currently displayed at  
http://www.bbc.com/sport/football/tables  
with individual clubs highlighted in API queries like this:  
http://push.api.bbci.co.uk/p?t=morph%3A%2F%2Fdata%2Fbbc-morph-sport-football-tables-data%2Fteam%2Farsenal%2FteamName%2FArsenal%2Fversion%2F1.4.1

--darkgray@synirc
