# webapp

An individual markdown file can be viewed as a _lesson_.
A folder of markdown files can be viewed as a _course_.

Example: A tutorial on _Benelux_, the politico-economic union
of the countries Belgium, the Netherlands and Luxembourg.

In such a tutorial, the first course could be about Benelux as a whole,
with lessons including a brief overview, a brief history and a
brief discussion of its economics.

This is followed by courses detailing Belgium, Netherlands, and Luxembourg.

Said courses may hold lessons on provinces, or sub-courses regional
histories, cities etc.  A user could drop in anywhere, but content should
be arranged such that a depth-first traversal of the hierarchy is a
meaningful path through all content. One ought to be able to read
it as a book.

Associated REST paths reflect the file system hierarchy, e.g.

    benelux.com/overview                   Describes Benelux in general.
    benelux.com/history                    Benelux history, economy, etc.
    benelux.com/economy
    benelux.com/belgium/overview           Describes Belgium in general.
    benelux.com/belgium/tintin             Dive into details.
    benelux.com/belgium/beer
    benelux.com/belgium/antwerp/overview   Dive into Antwerp, etc.
    benelux.com/belgium/antwerp/diamonds
    benelux.com/belgium/antwerp/rubens
    benelux.com/belgium/east-flanders
    benelux.com/belgium/brabant
    ...
    benelux.com/netherlands/overview
    benelux.com/netherlands/drenthe
    benelux.com/netherlands/flevoland
    ...

All content can be accessible from a left nav:

      overview     |                      |
      belgium      |                      |  {lesson
     [netherlands] |   {lesson content}   |   guide}
      luxembourg   |                      | 
                   |                      | 

At all times exactly one of the left nav choices is selected, and the
main page shows content associated with that selection.

The first item, in this case the "overview", is the initial highlight.
If one hits the domain without a REST path, one is redirected to
/overview and that item is highlighted in the menu, and its
content is shown.

Items in the left nav either name content and show it when clicked, or
they name sub-courses and expand choices when clicked.
In the latter case, the main content and the left nav highlighting
don't change.  Subsequent clicks toggle sub-course names.

Only the name of a lesson (a leaf) with content can 
 1) be highlighted,
 2) change the main page content when clicked, and
 3) serve at a meaningful REST address.

Everything else is a course, and only expands or hides
its own appearance.

This scheme maps to this filesystem layout:

     benelux/
       history.md
       economy.md
       README.md
       belgium/
         tintin.md
         beer.md
         antwerp/
           README.md
           diamonds.md
           ...
         east-flanders.md
         brabant.md
         ...
       netherlands/
         README.md
         drenthe.md
         flevoland.md
       ...

 Where, say README is converted to something labelled as `overview`
 by the UX. The ordering of files and directories in the tutorials
 is presented in, say, a file called `README_ORDER.txt` so that,
 for example, `history` may precede `economy`.

Useful data structures would facilitate mapping a string path, e.g.

> belgium/antwerp/diamonds
 
can map to some list of div ids to know what items to
show/hide when first loading a page.

Another handy structure would facilitate a prev/next navigation through
the lessons.  The lessons are all leaves of the tree.  Moving from lesson
_n_ to lesson _n+1_ means changing the main content and changing what is shown
or highlighted in the left nav.
