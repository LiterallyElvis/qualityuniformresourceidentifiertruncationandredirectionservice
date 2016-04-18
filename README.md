# qualityuniformresourceidentifiertruncationandredirectionservice

I thought it'd be funny to make a URL shortener that had a ridiculously long domain name

It exposes a very small number of routes where it accepts parameters thanks to [Gorilla/mux](https://github.com/gorilla/mux).

Then it uses [the Golang Markov chain generator example](https://golang.org/doc/codewalk/markov/) to generate a random string from the [Project Gutenberg eBook of Grimms' Fairy Tales](http://www.gutenberg.org/cache/epub/2591/pg2591.txt). It then enters that string into [Bolt DB](https://github.com/boltdb/bolt) as a key, with the submitted link as a value.