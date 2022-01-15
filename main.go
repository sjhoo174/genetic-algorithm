package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
	// "sort"
)

type child struct {
	mu           sync.Mutex
	data         []byte
	fitness      float64
	reprod_value int
}

type genetic struct {
	offspring  []child
	b1         float64
	b2         float64
	num        int
	cross_prob float64
}

func (gen *genetic) initialize() {
	// diff := gen.b2 - gen.b1
	var l int = 8
	// l = len(strconv.FormatInt(int64(diff), 2))/8 + 1
	gen.offspring = make([]child, gen.num)
	for i := 0; i < gen.num; i++ {
		b := make([]byte, l)
		rand.Read(b)
		c := child{
			data: b,
		}
		gen.offspring[i] = c

	}
}

func (gen *genetic) compute_population() {

	cross_array := make([]int, len(gen.offspring))
	for i := 0; i < len(gen.offspring); i++ {
		cross_array[i] = i
	}
	for i := range cross_array {
		j := rand.Intn(i + 1)
		cross_array[i], cross_array[j] = cross_array[j], cross_array[i]
	}
	q := make(chan int, len(gen.offspring))

	for i := 0; i < len(gen.offspring); i++ {
		
		target := cross_array[i]
		off := gen.offspring[i]
		cross_target := gen.offspring[target]
		go gen.crossover(off, cross_target, q)

		
	}
	for i := 0; i <len(gen.offspring); i++ {
		<-q
	}
	
	gen.mutate()
	gen.computefitness()
	gen.reproduce()
	

}

func (gen *genetic) crossover(first child, second child, q chan int) {
	
	first.mu.Lock()
	second.mu.Lock()

	defer first.mu.Unlock()
	defer second.mu.Unlock()
	r := rand.Intn(100)
	if r < 50 {
		q <- 1
		return
	}
	off := first.data
	cross_target := second.data
	len_bits := len(off) * 8
	cut := rand.Intn(len_bits)
	byte_index := cut / 8
	remainder := cut - byte_index*8

	// fmt.Println("Before crossover")
	// fmt.Println(off)
	// fmt.Println(cross_target)

	for i := 0; i < byte_index; i++ {
		temp := cross_target[i]
		cross_target[i] = off[i]
		off[i] = temp
	}
	for i := byte_index + 1; i < len(cross_target); i++ {
		temp := cross_target[i]
		cross_target[i] = off[i]
		off[i] = temp
	}
	left_off := (byte(255) << (8 - remainder)) & off[byte_index]
	right_off := (byte(255) >> remainder) & off[byte_index]
	left_cross := (byte(255) << (8 - remainder)) & cross_target[byte_index]
	right_cross := (byte(255) >> remainder) & cross_target[byte_index]

	off[byte_index] = left_cross | right_off
	cross_target[byte_index] = left_off | right_cross
	q <- 1

	// fmt.Println("After crossover")
	// fmt.Println(off)
	// fmt.Println(cross_target)

}

func (gen *genetic) computefitness() {
	// fmt.Println("Fitness")
	// for i := range gen.offspring {
	// 	fmt.Println(gen.offspring[i].data)
	// }

	for i := range gen.offspring {
		var sum uint32 = 0
		child := &gen.offspring[i]
		
		child.mu.Lock()
		for j := len(child.data) - 1; j >= 0; j-- {
			sum |= (uint32(child.data[j]) << ((-j + len(child.data) - 1) * 8))
		}
		var diff uint32 = 0
		var count int = 0
		for count <= 63 {
			diff ^= (1 << count)
			count += 1
		}

		fitscore := gen.b1 + (float64(sum) / float64(diff) * (gen.b2-gen.b1))
		fitscore = math.Pow(float64(fitscore), 2)
		
		child.fitness = fitscore
		child.mu.Unlock()

	}
	// for i := range gen.offspring {
	// 	fmt.Println(gen.offspring[i].fitness)
	// }

	

}

func (gen *genetic) reproduce() {
	var avg float64
	for z := range gen.offspring {
		avg += gen.offspring[z].fitness
	}
	avg /= float64(gen.num)
	// for z:= range gen.offspring {
	// 	fmt.Println(gen.offspring[z].data)
	// }
	// for z:= range gen.offspring {
	// 	fmt.Println(gen.offspring[z].fitness)
	// }
	for z := range gen.offspring {
		gen.offspring[z].fitness /= avg
		gen.offspring[z].reprod_value = int(math.Round(gen.offspring[z].fitness))
	}

	
	// for z:= range gen.offspring {
	// 	fmt.Println(gen.offspring[z].reprod_value)
	// }
	newoffspring := []child{}
	for z := range gen.offspring {
		for i := 0; i < gen.offspring[z].reprod_value; i++ {
			newoffspring = append(newoffspring, gen.offspring[z])
		}
	}
	// fmt.Println("before")
	// for i := range gen.offspring {
	// 	fmt.Println(gen.offspring[i].data)
	// }
	// sort.SliceStable(newoffspring, func (i,j int) bool {
	// 	return newoffspring[i].fitness > newoffspring[j].fitness
	// })
	gen.offspring = newoffspring
	// fmt.Println(len(gen.offspring))
	// fmt.Println("after")
	// for i := range gen.offspring {
	// 	fmt.Println(gen.offspring[i].data)
	// }

}
func (gen *genetic) mutate() {

	for i := range gen.offspring {
		child := &gen.offspring[i]
		for j := range child.data {
			for k := 0; k < 8; k++ {
				ran := rand.Intn(1000)
				if ran > 999 {
					child.data[j] ^= 1 << k
				}
			}
		}
		
	}
	
}

func getParameters(w http.ResponseWriter, req *http.Request) {
	var bone float64
	var btwo float64
	var num_off int
	var cross_probability float64
	var _ error

	if b_one_keys, e1 := req.URL.Query()["b1"]; e1 == true{
		bone, _ = strconv.ParseFloat(b_one_keys[0], 64)

	}
	if b_two_keys, e2 := req.URL.Query()["b2"]; e2 == true {
		btwo, _ = strconv.ParseFloat(b_two_keys[0], 64)

	}
	if num_keys, e3 := req.URL.Query()["num"]; e3 == true {
		num_off, _ = strconv.Atoi(num_keys[0])

	}
	if cross_prob_keys, e4 := req.URL.Query()["cross_prob"]; e4 == true {
		cross_probability, _ = strconv.ParseFloat(cross_prob_keys[0], 64)

	}
	var g genetic

	g = genetic{
		b1:         bone, //-50
		b2:         btwo, //50
		num:        num_off, //50
 		cross_prob: cross_probability, //0.5
	}
	start := time.Now()
	g.initialize()
	for i := 0; i <10000; i++ {
		fmt.Fprintf(w, "%v\n", i)
		g.compute_population()
		
	}
	elapsed := time.Since(start)

	for i := range g.offspring {
		fmt.Fprintf(w, "%v\n", g.offspring[i].data)
	}
	fmt.Fprintf(w, "Elapsed time %s", elapsed)
}

func main() {

	http.HandleFunc("/param", getParameters)
	http.ListenAndServe("0.0.0.0:8080", nil)
	
	

}
