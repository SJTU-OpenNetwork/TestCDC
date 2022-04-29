package main

import (
	"fmt"
	"os"
    "io"
    "bufio"
    "time"
    "strconv"
    chunker "github.com/SJTU-OpenNetwork/go-ipfs-chunker"
    blocks "github.com/ipfs/go-block-format"
)

func chunkData(s chunker.Splitter) (map[string]blocks.Block, error) {

    
    blkmap := make(map[string]blocks.Block)

    for {
        blk, err := s.NextBytes()
        if err != nil {
            if err == io.EOF {
                break
            }
            return blkmap, err
        }

        b := blocks.NewBlock(blk)
        blkmap[b.Cid().KeyString()] = b
    }

    return blkmap, nil
}


func diff(file1 string, file2 string, alg string, min int, avg int, max int) {
    fi1,err := os.Open(file1)
    if err != nil{panic(err)}
    defer fi1.Close()

    fi2,err := os.Open(file2)
    if err != nil{panic(err)}
    defer fi2.Close()


    var newCDC func(r io.Reader) chunker.Splitter

    switch alg{
    case "HRAM":
        newCDC = func(r io.Reader) chunker.Splitter {
            return chunker.NewHram(r, min, avg, max, 8)
        }
    case "FastCDC":
        newCDC = func(r io.Reader) chunker.Splitter {
            return chunker.NewFastCDC(r, uint64(min), uint64(avg), uint64(max))
        }
    case "Rabin":
        newCDC = func(r io.Reader) chunker.Splitter {
            return chunker.NewRabin(r, uint64(avg))
        }
    case "RAM":
        newCDC = func(r io.Reader) chunker.Splitter {
            return chunker.NewRam(r, min, max, 4)
        }
    default:
        return
    }

    s1 := newCDC(bufio.NewReader(fi1))
    s2 := newCDC(bufio.NewReader(fi2))

    start := time.Now()
    blk1,_ := chunkData(s1)
    blk2,_ := chunkData(s2)
    end := time.Now()


    var extraCount, extraSize int
	for k := range blk2 {
		_, ok := blk1[k]
		if !ok {
			extraCount++
            extraSize+=len(blk2[k].RawData())
		}
	}

    f1stat, err := fi1.Stat()
    f1size := f1stat.Size()

    f2stat, err := fi2.Stat()
    f2size := f2stat.Size()
    p := float64(extraSize)/float64(f2size)

    elapse := end.Sub(start)
    fmt.Println(f1size+f2size)
    fmt.Println(elapse.Seconds())
    throughput := float64(f1size+f2size)/float64(elapse.Microseconds())

    fmt.Printf("Throughput: %f Mbps, ExtraCount: %d, ExtraSize: %d, ExtraP: %f\n", throughput, extraCount, extraSize, p)
    
}

func printHelp() {
    fmt.Println("USAGE: ./main file1 file2 algorithm args(min/avg/max)")
}


func main() {

    if len(os.Args) != 7 {
        printHelp()
        return
    }

    min, err := strconv.Atoi(os.Args[4])
	if err != nil{
		panic(err)
    }

    avg, err := strconv.Atoi(os.Args[5])
	if err != nil{
		panic(err)
    }

    max, err := strconv.Atoi(os.Args[6])
	if err != nil{
		panic(err)
    }


    diff(os.Args[1], os.Args[2], os.Args[3], min, avg, max)
}
