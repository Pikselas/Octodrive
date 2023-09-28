# Octodrive
Beloved Octocat as Personal backup storage ??

```go
// create user
octoUser, octo_err := ToOcto.NewOctoUser(
  "[[ YOUR EMAIL HERE]]",
  "[[ ACCESS TOKEN HERE ]]")

if octo_err != nil {
  panic(octo_err)
}

// create drive (root path as default)
octoDrive, err := Octo.NewOctoDrive(octoUser, Octo.DefaultFileRegistry)
if err != nil {
 panic(err)
}

// open a file
file , err := os.Open("your_file.extension")
if err != nil {
  panic(err)
}

o_file := octoDrive.Create(file)

// optional file configuration
// [ i.e encryption of file | compression of file | etc... ]

// write chunk by chunk or all at once
err = o_file.WriteAll

//handle if any error occurs
if err != nil {
  err = 	o_file.RetryWriteChunk()
}

// repeat retrying
// -----------------

// Save the file
err = octoDrive.Save("file_path/file_name" , o_file)

```
