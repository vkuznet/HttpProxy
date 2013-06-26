HttpProxy
=========

One day I realized that I need a proxy server to constrain my kids from
wildness of the internet. And the project was born. Since I love programming in
Go I decided to give it a shot. Moreover I found excellent goproxy [1] package
which did almost all of the work. The HttpProxy package supports white and
black lists as well as more flexible rule list, see below. I hope you'll find
it useful.

White/black lists
-----------------
HttpProxy supports white and black lists. Eeach of them can be specified in
separate files, e.g. whitelist.txt and blacklist.txt. The content of those
files is a list of sites you want to have, e.g.

```
google.com
amazon.com
```

Please note that HttpProxy will use pattern as is, therefore if site has
multiple domains it is better to use its base URL address, e.g. amazon.com.
But due to "as is" nature of those lists you can pass any regular expression
patterns, e.g.

```
^www.amazon.com$
```

which stands for site which always starts with www and ends with com.

Rule list
---------
Suppose you want to restrict access to certain sites with some policy, e.g.
only between 9am and noon. To do so create a rules.txt file with the following
content

```
www.facebook.com,9,12
www.myspace.com,12,15
```

The HttpProxy will read its content and apply this rules to the proxy.

Usage
-----
To build the executable just run
```
go build
```

To run it, you may invoke it from your command line or use run.sh script
Finally, you'll need to configure your browser accordingly to use the proxy.

References
----------
[1] github.com/elazarl/goproxy

License
-------
The code is released under BSD license.
