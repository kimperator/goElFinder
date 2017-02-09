package elFinder

/*
type volumeParents struct {
	Tree []volumeFileDir `json:"tree"`
}


type volumeDriver interface {
	name() string
	root() string

	isDir(path string) bool
	isFile(path string) bool
	fileInfo(path string) volumeFileDir
	listDir(path string) ([]string, error)
	file(path string) (string, []byte, error)
	mkdir(target, name string) (error)
}
*/

type volumeConfig struct {
	init bool
	tree bool
	rootDir string // [name]realPath
	dirsRight map[string]bool
	defaultRight bool
}

type volumeResponse struct {
	config volumeConfig

	Api string `json:"api,omitempty"` // The version number of the protocol, must be >= 2.1, ATTENTION - return api ONLY for init request!
	Cwd volumeFileDir `json:"cwd,omitempty"` // Current Working Directory - information about the current directory. Information about File/Directory
	Files []volumeFileDir `json:"files,omitempty"` // array of objects - files and directories in current directory. If parameter tree == true, then added to the folder of the directory tree to a given depth. The order of files is not important. Note you must include the top-level volume objects here as well (i.e. cwd is repeated here, in addition to other volumes)
	NetDrivers []string `json:"netDrivers,omitempty"` // Network protocols list which can be mounted on the fly (using netmount command). Now only ftp supported.
	Options volumeOptions `json:"options,omitempty"`
	UplMaxFile string `json:"uplMaxFile,omitempty"` // Allowed upload max number of file per request. For example 20
	UplMaxSize string `json:"uplMaxSize,omitempty"` // Allowed upload max size per request. For example "32M"

	Dim string `json:"dim,omitempty"` // for images

	Added []volumeFileDir `json:"added,omitempty"` // for upload, mkdir, rename
	Warning []string `json:"warning,omitempty"` // for upload
	Changed []volumeFileDir `json:"changed,omitempty"` // for mkdir
	Hashes map[string]string `json:"hashes,omitempty"` // for mkdir

	Removed []string `json:"removed,omitempty"` // for remove, rename

	Error interface{} `json:"error,omitempty"`
}

type volumeOptions struct {
	Path string `json:"path,omitempty"` // Current folder path
	Url string `json:"url,omitempty"` // Current folder URL
	TmbUrl string `json:"tmbURL,omitempty"` // Thumbnails folder URL
	Separator string `json:"separator,omitempty"` // Path separator for the current volume
	Disabled []string `json:"disabled,omitempty"`  // List of commands not allowed (disabled) on this volume
	// ToDo https://github.com/Studio-42/elFinder/wiki/Client-Server-API-2.1#open

}

type volumeFileDir struct {
	Name string `json:"name,omitempty"` // name of file/dir. Required
	Hash string `json:"hash,omitempty"` //  hash of current file/dir path, first symbol must be letter, symbols before _underline_ - volume id, Required.
	Phash string `json:"phash,omitempty"` // hash of parent directory. Required except roots dirs.
	Mime string `json:"mime,omitempty"` // mime type. Required.
	Ts int64 `json:"ts,omitempty"` // file modification time in unix timestamp. Required.
	Size int64 `json:"size,omitempty"` // file size in bytes
	Dirs byte `json:"dirs,omitempty"` // Only for directories. Marks if directory has child directories inside it. 0 (or not set) - no, 1 - yes. Do not need to calculate amount.
	Read byte `json:"read,omitempty"` // is readable
	Write byte `json:"write,omitempty"` // is writable
	Isroot byte `json:"isroot,omitempty"`
	Locked byte `json:"locked,omitempty"` // is file locked. If locked that object cannot be deleted, renamed or moved
	Tmb string `json:"tmb,omitempty"` // Only for images. Thumbnail file name, if file do not have thumbnail yet, but it can be generated than it must have value "1"
	Alias string `json:"alias,omitempty"` // For symlinks only. Symlink target path.
	Thash string `json:"thash,omitempty"` // For symlinks only. Symlink target hash.
	Dim string `json:"dim,omitempty"` // For images - file dimensions. Optionally.
	Isowner bool `json:"isowner,omitempty"` // has ownership. Optionally.
	Cssclr string `json:"cssclr,omitempty"` // CSS class name for holder icon. Optionally. It can include to options.
	Volumeid string `json:"volumeid,omitempty"` // Volume id. For directory only. It can include to options.
	Netkey string `json:"netkey,omitempty"` // Netmount volume unique key, Required for netmount volume. It can include to options.
//	Options volumeOptions `json:"options,omitempty"` // For volume root only. This value is same to cwd.options.
}

