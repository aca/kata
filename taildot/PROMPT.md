
Write simple cli tool that works like 'tail -f' but if there's no data it inserts '.' so it makes it much convenient to analyze log.
It should print log as fast as it can, to debug

// if program outputs 
10:00:00 hello
10:00:01 world
10:01:00 hello

// it converts it to 
10:00:00 hello
10:00:01 world
.
.
.
10:01:00 hello

