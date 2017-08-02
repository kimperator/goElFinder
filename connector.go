/*
ElFinder connector handler

Example code:
	config := elFinder.Config{}
	config["l0"] = elFinder.Volume {
		Root: "./files/1",
		AllowDirs: []string{"/Allow"},
		DenyDirs:  []string{"/Deny"},
		DefaultRight: false,
	}
	config["l1"] = elFinder.Volume {
		Root: "./files/2",
		DefaultRight: true,
	}
	mux.Handle("/connector", elFinder.NetHttp(config))
*/
package goElFinder

import (
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"github.com/go-playground/form"
)

const APIver = "2.1"

var decoder *form.Decoder

func (volumes Volumes) NetHttp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err error
		)
		//conf = config

// ToDo use it--------------------ssi
		var self elf
		self.volumes = volumes

		decoder = form.NewDecoder()

//--------------------

		if r.Method == "GET" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = decoder.Decode(&self.req, r.Form)
			if err != nil {
				log.Println(err)
			}


		} else if r.Method == "POST" {
			r.ParseMultipartForm(32 << 20) // ToDo check 8Mb

			err = decoder.Decode(&self.req, r.PostForm)
			if err != nil {
				log.Println(err)
			}
		}
//-------------------------------------------------------------------------

		err = self.parse()
		self.target, err = self.volumes.parsePathHash(self.req.Target)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"error" : "%s"}`, err)))
			return
		}

//-------------------------------------------------------------------------

		switch self.req.Cmd {
		case "open":
			err := self.open()
			if err != nil {
				log.Println("Volume open:", err)
			}

		case "file":
			fileName, mimeType, data, err := self.file()
			if err != nil {
				self.res.Error = err.Error()
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

		case "tree":
			err := self.tree(self.target)
			if err != nil {
				self.res.Error = err.Error()
			}
		case "parents":
			err := self.parents(self.target)
			if err != nil {
				self.res.Error = err.Error()
			}
		case "ls":
			self.ls()
		case "tmb":
			self.tmb()

		case "size":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"size": %d}`, self.size())))
			return
		case "dim":
			err := self.dim()
			if err != nil {
				self.res.Error = err.Error()
			}
		case "mkdir":
			if len(self.req.Dirs) > 0 { // ToDo this
				self.mkdirs()
			} else {
				err = self.mkdir()
				if err != nil {
					self.res.Error = err.Error()
				}
			}
		case "mkfile":
			err = self.mkfile()
			if err != nil {
				self.res.Error = err.Error()
			}
		case "rm":
			err = self.rm()
			if err != nil {
				self.res.Error = err.Error()
			}
		case "rename":
			err := self.rename(self.target.id, self.target.path)
			if err != nil {
				self.res.Error = err.Error()
			}

		case "duplicate":
			//ToDo
		case "paste":
			self.paste()
		case "upload":

			if self.req.Chunk != "" {
				if r.PostForm["cid"] == nil {
					if len(self.req.Renames) != 0 {
						fmt.Println("Result renames", self.renames(self.target.id, self.target.path))
					}
					err = self.chunkMerge(self.target.id, self.uploadpath[0].path, self.req.Chunk)
					if err != nil {
						self.res.Error = err
					}
				} else {
					for i := range r.MultipartForm.File["upload[]"] {
						file, err := r.MultipartForm.File["upload[]"][i].Open()
						if err != nil {
							fmt.Println(err)
						}
						err = self.chunkUpload(self.target.id, self.uploadpath[i].path, self.req.Chunk, file)
						if err != nil {
							self.res.Error = err
						}
					}
				}
			} else {
				if len(self.req.Renames) != 0 {
					fmt.Println("Result renames", self. renames(self.target.id, self.target.path))
				}
				ers := []string{}
				for i, f := range r.MultipartForm.File["upload[]"] {
					file, _ := f.Open()
					er := self.upload(self.target.id, self.uploadpath[i].path, f.Filename, file)
					if er != nil {
						ers = append(ers, er.Error())
					}
					self.req.Renames = []string{}
				}
				if len(ers) > 0 {
					self.res.Error = ers
				}
			}


		case "get":
			err = self.get()
			if err != nil {
				self.res.Error = err.Error()
			}
		case "put":
			err = self.put()
			if err != nil {
				self.res.Error = err.Error()
			}
		case "archive":
			//ToDo
		case "extract":
			//ToDo
		case "search":
			//ToDo
		case "info":
			//ToDo
		case "resize":
			switch self.req.Mode {
			case "resize":
				err = self.resize(self.target.id, self.target.path)
			case "crop":
				err = self.crop(self.target.id, self.target.path)
			case "rotate":
				err = self.rotate(self.target.id, self.target.path)
			}
			if err != nil {
				self.res.Error = err.Error()
			}
		case "url":
			self.url()
		//	case "netmount":
		case "zipdl":
			//ToDo
		case "callback":
			//ToDo
		case "chmod":
			//ToDo


		default:
			self.res.Error = "errUnknownCmd"

		}

		js, err := json.Marshal(self.res)
		if err != nil {
			js = []byte(`{"error" : ["errConf", "errJSON"]}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	})
}
