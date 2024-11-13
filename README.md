<div align="center">
  <h1>
    <small>sano</small><br>sitegen
  </h1>
</div>

## what?

a small personal site generator that is way too opinated and small for it to
make sense for others to use it. i wanted something that could turn my `.md`
files to `.html` files. so, this is that glue code that turns my markdown to
HTML and does few more things that i want in my site.

## why?

after experimenting with several static site generators, i reached a point where
none of the solutions seemed to fit me (for more [read this](https://sudanchapagain.com.np/writings/writing-html-is-hard)).
everything came with their own quirks, and it just didn’t feel like it was made
for me and my use case. so, i decided build my own solution because my
requirement was as simple as it gets. it just turns markdown files into HTML.
there’s no complex setup, no features, just glue code that gets the job done.

## how?

markdown file with following frontmatter supports (title, desc, date, css, js,
status). state decides if that file should be public or not. css & js are
specific to that page. all asset's are assumed to be inside `/src/assets/`.
CSS and JS file's path are assumed to be `/src/assets/css` and `/src/assets/js`
respectively.

everything should be inside `/src/` and output is generated inside of `/dist/`.
since, `/src/assets/` is copied in whole to `/dist/`, version control and
deployment strategy should be considered to not have duplicates in two different
folders.

example directory structure:

```
├───dist/
│   ├───css/
│   ├───img/
│   ├───js/
│   ├───index.html
│   └───writings/
└───src/
    ├───assets
    │   ├───css/
    │   ├───img/
    │   └───js/
    ├───index.md
    └───writings/

```
