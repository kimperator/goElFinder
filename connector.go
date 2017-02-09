package elFinder

import (
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
)

const APIver = "2.1"


func NetHttp(w http.ResponseWriter, r *http.Request)  {
	var (
		cmd string
		target string
		targets []string
	)
	volume := new(volumeResponse)

// Test data config
	volume.setRoot("./test")
	volume.allowDirs([]string{"/1"})
	volume.setDefaultRight(false)
// /Test data config

	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
fmt.Println("GET:", r.Form)
		if r.Form["init"] != nil && r.Form["init"][0] == "1" {
			volume.config.init = true
		}
		if r.Form["tree"] != nil && r.Form["tree"][0] == "1" {
			volume.config.tree = true
		}

		if r.Form["target"] != nil {
			_, target, _ = parseHash(r.Form["target"][0])

		} else if r.Form["targets[]"] != nil {
			for _, ft := range r.Form["targets[]"] {
				_, p, _ := parseHash(ft)
				targets = append(targets, p)
			}
		} else {
			target = "/"
			return
		}

		if r.Form["cmd"] == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error" : "errUnknownCmd"}`))
			return
		}
		cmd = r.Form["cmd"][0]
	} else if r.Method =="POST" {
		r.ParseMultipartForm(32 << 20)
fmt.Println("POST", r.PostForm)
		if r.PostForm["target"] != nil {
			_, target, _ = parseHash(r.PostForm["target"][0])

		} else if r.PostForm["targets[]"] != nil {
			for _, ft := range r.PostForm["targets[]"] {
				_, p, _ := parseHash(ft)
				targets = append(targets, p)
			}
		} else {
			target = string(filepath.Separator)
			return
		}
		cmd = r.PostForm["cmd"][0]
	}



//-------------------------------------------------------------------------
	switch cmd {
	case "open":
		if volume.checkRight(target) {
			err := volume.open(target)
			if err != nil {
				log.Println(err)
			}

		} else {
			volume.Error = []string{"errLocked", target,}
		}

	case "file":
		if volume.checkRight(target) {
			fileName, mimeType, data, err := volume.file(target)
			if err != nil {
				volume.Error = err.Error()
			} else {
				w.Header().Set("Content-Type", mimeType)
				if r.Form["download"] != nil {
					w.Header().Set("Content-Disposition", "attachment; filename='" + fileName + "'")
				} else {
					w.Header().Set("Content-Disposition", "inline; filename='" + fileName + "'")
				}
				w.Write(data)
				return
			}
		} else {
			volume.Error = []string{"errLocked", target,}
		}

	case "tree":
	case "parents":
	case "ls":
	case "tmb":
	case "size":

	case "dim":
		if volume.checkRight(target) {
			err := volume.dim(target)
			if err != nil {
				volume.Error = err.Error()
			}
		} else {
			_, path, _ := parseHash(target)
			volume.Error = []string{"errLocked", path,}
		}

	case "mkdir":
		if r.Form["name"] != nil {
			fmt.Println(volume.mkdir(target, r.Form["name"][0])) // ToDo
		}
		if len(r.Form["dirs[]"]) > 0 {
			err := []string{}
			for _, f := range r.Form["dirs[]"] {
				fmt.Println("Make dir:",f)
				e := volume.mkdir(target, f)
				if e != nil {
					err = append(err, e.Error())
				}
			}
			if len(err) > 0 {
				volume.Error = err
			}
		}

	case "mkfile":
	case "rm":
		err := []string{}
		for _, f := range targets {
			e := volume.rm(f)
			if e != nil {
				err = append(err, e.Error())
			}
		}
		if len(err) > 0 {
			volume.Error = err
		}

	case "rename":
		if r.Form["name"] != nil {
			err := volume.rename(target, r.Form["name"][0])
			if err != nil {
				volume.Error = err.Error()
			}
		}
		fmt.Println(volume)

	case "duplicate":
	case "paste":
	case "upload":
		err := []string{}
		for i, f := range r.MultipartForm.File["upload[]"] {
			fmt.Println(r.PostForm["upload_path[]"][i],":", f.Filename)
			file, _ := f.Open()
			_, path, _ := parseHash(r.PostForm["upload_path[]"][i])
			e := volume.upload(path, f.Filename, file)
			if e != nil {
				err = append(err, e.Error())
			}
		}
		if len(err) > 0 {
			volume.Error = err
		}

	case "get":
	case "put":
	case "archive":
	case "extract":
	case "search":
	case "info":
	case "resize":
	case "url":
//	case "netmount":
	case "zipdl":
	case "callback":
	case "chmod":


	default:
		volume.Error = "errUnknownCmd"

	}

	js, err := json.Marshal(volume)
	if err != nil {
		js = []byte(`{"error" : ["errConf", "errJSON"]}`)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
