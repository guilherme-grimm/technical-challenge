package entity

// entities live here, easy to access, keep things clean and smol files

// add device def
type Device struct {
}

type PaginatedResponse struct {
	data   []Device
	cursor string
}
