A default blog scaffold for **ruhoh** compatible static blog parsers.

**This version is for [ruhohSpec v1.0](http://ruhoh.com/universal-blog-api/) and is only compatible with ruhoh gem 1.0.x.x**

See <http://ruhoh.com> for more information.

## Usage and Documentation

See <http://ruhoh.com/usage> for full usage and documentation.

Currently the only parser is available in ruby as a ruby gem:

````bash
$ gem install ruhoh --version '1.0.0.alpha'
$ ruhoh help
````

Once you have the gem you can run your blog locally:

Assume USERNAME is your GitHub username.

````bash
$ git clone git://github.com/ruhoh/blog.git USERNAME.ruhoh.com
$ cd USERNAME.ruhoh.com
$ rackup -p 9292
````

Using your GitHub username is useful for automatically publishing to ruhoh.com

For all publishing options including self-hosting see: <http://ruhoh.com/usage/publish>

## License

Released under the [MIT License](http://www.opensource.org/licenses/MIT)

