# Loader

This package loads markdown files from a file system into memory,
retaining the file system tree structure.

```
ldr := loader.NewFsLoader(afero.NewOsFs())
folder, err := ldr.LoadTrees(args)
if err != nil {
  return err
}
visitor.VisitFolder(folder)
```

`LoadTrees` returns an instance of `MyFolder`.

`MyFolder` holds slices of `MyFile` and `MyFolder`.

Each of these is a _tree node_ with a `Name`, a full `Path`,
and a visitor acceptance method.
