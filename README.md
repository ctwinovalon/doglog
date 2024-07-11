# doglog

Doglog is a cli tool for searching logs in Datadog and pretty-printing them
to the console. It requires a Datadog app key to be defined for the user.

The output is controlled by user-defined formats. The formats are defined in
order from most-specific to least-specific:

```
java1 = {{.__timestamp}} {{.__service | printf "%-20.20s"}} {{._Level_color}}{{.__level | printf "%-5.5s"}}{{._Reset}} [{{.__threadname | printf "%-20.20s"}}] {{.__short_classname | printf "%-30.30s"}} -- {{._Level_color}}{{.__message}}{{._Reset}}
java2 = {{.__timestamp}} {{.__service | printf "%-20.20s"}} {{._Level_color}}{{.__level | printf "%-5.5s"}}{{._Reset}} {{.__short_classname | printf "%-30.30s"}} -- {{._Level_color}}{{.__message}}{{._Reset}}
java3 = {{.__timestamp}} {{.__service | printf "%-20.20s"}} {{._Level_color}}{{.__level | printf "%-5.5s"}}{{._Reset}} [{{.__threadname | printf "%-20.20s"}}] {{._Level_color}}{{.__message}}{{._Reset}}
minimal = {{.__timestamp}} {{.__service | printf "%-20.20s"}} {{._Level_color}}{{.__level | printf "%-5.5s"}} {{.__message}}{{._Reset}}
```

The formats use Go [language templates](https://pkg.go.dev/text/template). These templates
are supplied with the fields from Datadog. By providing both long and short formats,
users can switch from the command-line between which to use.
The names of the formats don't matter, they just have to be unique.

The formats, api keys, etc. are defined in a configuration file named `.doglog`. By default this file
is defined in the user's home directory. You can run `doglog --help` and look at the
config file argument to see where `doglog` expects to find it.

You can review an [example configuration file](https://raw.githubusercontent.com/ctwinovalon/doglog/main/example.doglog).

In addition to the "normal" Go language template functions, the [Sprig functions](https://masterminds.github.io/sprig/)
can also be used in the template definitions.

There are a number of options for `doglog` but only one is required, the `-s, --service`
argument. The service argument is required to constrain the log
searches.

```man
usage: datadog [-h|--help] [-c|--config "<value>"] [-d|--debug] [-i|--indices
               "<value>" [-i|--indices "<value>" ...]] [-j|--json] [-l|--limit
               <integer>] [--long] [--no-colors] [-q|--query "<value>"]
               -s|--service "<value>" [--start "<value>"] [--end "<value>"]
               [-t|--tail] [-v|--version]

               Search and tail logs from Datadog.

Arguments:

  -h  --help       Print help information
  -c  --config     Path to the config file. Default: /home/ctwise/.doglog
  -d  --debug      Generate debug output.
  -i  --indices    The list of indices to search in Datadog. Repeat the
                   parameter to add indices to the list. Default: [main]
  -j  --json       Output messages in json format. Shows the modified log
                   message, not the untouched message from Datadog. Useful in
                   understanding the fields available when creating Format
                   templates or for further processing.
  -l  --limit      The maximum number of messages to request from Datadog. Must
                   be greater then 0. Default: 300
      --long       Generate long output. Default: false
      --no-colors  Don't use colors in output. Automatically turned off when
                   redirecting output.
  -q  --query      Query terms to search on (Datadog search syntax). Defaults
                   to '*'.
  -s  --service    The Datadog log 'service' to constrain the log search, e.g.,
                   '-s send-email'.
      --start      Starting date/time to search from. The start and end
                   parameters can be: 1) an ISO-8601 string using the FULL
                   format of '2024-07-11T08:45:00+00:00', 2) a unix timestamp
                   (number representing the elapsed milliseconds since epoch),
                   3) a date math string such as +1h to add one hour, -2d to
                   subtract two days, etc. The full list includes s for
                   seconds, m for minutes, h for hours, and d for days.
                   Optionally, use now to indicate current time. Default:
                   now-15m
      --end        Ending date/time to search from. Uses Datadog format.
                   Defaults to 'now' if --start is provided but no --end.
                   Default: now
  -t  --tail       Whether to tail the output. Requires a relative search.
  -v  --version    Display the application version and exit.
```

Examples:
```
Tail the uis-api service
> doglog -s uis-api -t

Tail the uis-api service starting from 5 minutes ago
> doglog -s uis-api -t --start "now-5m"

Display the log messages between 3pm and 4pm
> doglog -s uis-api --start "2024-07-11T15:00:00+00:00" --end "2024-07-11T16:00:00+00:00"
```