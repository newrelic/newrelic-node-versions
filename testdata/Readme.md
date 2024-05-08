## bare-repo.git

This is a bare Git repository used for testing repo cloning. To add/remove
files in it:

```sh
$ git clone bare-repo.git foo-repo
$ cd foo-repo
# make changes
$ git add .
$ git commit -m 'whatever'
$ cd .. && rm -rf foo-repo
```
