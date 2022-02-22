# myWings-mirrorer

## What
A tool to automatically download all the files of myWings your user has permissions to.

### Features
- Only download new or changed files. Don't download files already downloaded.
- Ability to call a script/tool after a file was downloaded.

## Why
Die myWings webseite, sowie App haben einige Limitierungen
- Man bekommt nicht mit wenn neue Dokumente hochgeladen werden / bestehende upgedated werden.
- Man kann nur von geräten Zugreifen die einen Browser besitzen oder auf Android oder IOS basieren. Mit ebook Readern wirds schwer / unnötig kompliziert.
- Man kann nichts automatisiern. Z.B. alle PDF files automatisch nach dem Download durch ein Skript jagen.
- Weil ichs kann.

## How
    Usage:
      myWings-mirrorer [OPTIONS]
    
    Application Options:
      -u, --user=        The username of wings
      -p, --password=    The users password
      -o, --filename=    The filename used for downloding files (optional)
      -c, --cache=       The cachefile to be used (optional)
      -m, --postcommand= A command to run after the file was downloaded. (optional)
    
    Help Options:
      -h, --help         Show this help message

### Example
    myWings-mirrorer -u st123456 -p mypassword -m removeWatermarkPdf

### Where
The files are saved in the directory/format given to filename.
It defaults to ~/documents/%programName%/%courseName%/%fileTitle%

The names in %% are automatically replaced.
so a resolved download path looks like this:

    /home/chris/documents/Master IT-Sicherheit und Forensik/Einführung in die IT-Sicherheit und Forensik/Studienbrief_ Einführung in die IT-Sicherheit und Forensik.pdf

Chars known to make problems e.g. : are replaced with an underscore.

## When
The idea is to let this tool be run regularily by a cron job, so it automatically mirrors the content to the local drive. 
