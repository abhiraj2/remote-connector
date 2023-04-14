package idkeeper

import (
	"math/rand"
	"sync"
)

type Idkeeper struct{
	mu sync.Mutex
	id_arr []uint16
	rand_gen *rand.Rand
	last_idx int
} 

func (idk *Idkeeper) Init(){
	idk.mu.Lock()
	defer idk.mu.Unlock()
	idk.id_arr = make([]uint16, 1024)
	idk.last_idx = 0
	newSrc := rand.NewSource(42)
	idk.rand_gen = rand.New(newSrc)
	//idk.rand_gen.Seed(42)
	//fmt.Println(idk.rand_gen.Int())
}

func check_notin_array(arr []uint16, new_num uint16) bool{
	for _, ele := range arr{
		if new_num == ele {
			return false
		}
	}
	return true
}

func remove(s []uint16, ele uint16) []uint16{
	var res []uint16
	for _, element := range s{
		if element != ele{
			res = append(res, element)
		}
	}
	return res
}

func (idk *Idkeeper) AddElem() (uint16, bool){
	idk.mu.Lock()
	defer idk.mu.Unlock()
	unique := false
	new_num := uint16(idk.rand_gen.Int())
	for unique == false {
		unique = check_notin_array(idk.id_arr[:], uint16(new_num))
		new_num = uint16(idk.rand_gen.Int())
	}
	idk.id_arr[idk.last_idx] = new_num
	idk.last_idx++
	return new_num, false
}

func (idk *Idkeeper) RemoveElem(id uint16) error{
	idk.mu.Lock()
	defer idk.mu.Unlock()
	idk.id_arr = remove(idk.id_arr, id)
	return nil
}