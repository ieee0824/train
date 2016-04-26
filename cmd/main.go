package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

var docDir = "/var/www/html"

func upload(w http.ResponseWriter, r *http.Request) {
	form := `
 <html>
<title>Go upload</title>
<body>
<form action="`
	form += `/deploy" method="post" enctype="multipart/form-data">
<label for="file">Filename:</label>
<input type="file" name="file" id="file">
<input type="submit" name="submit" value="Submit">
</form>

</body>
</html>
  `
	w.Write([]byte(form))
}

func deploy(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	out, err := os.Create("/tmp/" + header.Filename)
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	err = Unzip("/tmp/"+header.Filename, docDir)
	if err == nil {
		fmt.Fprintf(w, "File uploaded successfully : ")
		fmt.Fprintf(w, header.Filename)
	} else {
		if bin, err := ioutil.ReadFile("/tmp/" + header.Filename); err == nil {
			err = ioutil.WriteFile(docDir+"/"+header.Filename, bin, 0644)
			if err == nil {
				fmt.Fprintf(w, "File uploaded successfully : ")
				fmt.Fprintf(w, header.Filename)
				return
			}
			fmt.Fprintf(w, err.Error()+"\n")
		}
	}
	fmt.Fprintf(w, "File uploaded failed.")
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getIP() (ip string) {
	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}

	for _, a := range addrs {
		ip = a
		return
	}
	return
}

func main() {

	http.HandleFunc("/deploy", deploy)
	http.HandleFunc("/", upload)
	http.ListenAndServe(":8080", nil)
}
