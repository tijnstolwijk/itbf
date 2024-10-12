package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

func main() {
  in := os.Args[1:][0]
  out := os.Args[1:][1]
  width, err := strconv.Atoi(os.Args[1:][2])
  if err != nil {
    fmt.Println("The string parameter must be a number")
    panic(err)
  }

  //Note: removed spanish Ñ
  characters := []rune("@#W$9876543210?!abc;:+=-,._   ") 

  // Open our input file
  reader, err := os.Open(in)
  if err != nil {
    fmt.Println("Error reading your file, did you do it correctly?")
    panic(err)
  }

  // Decode jpg to Image
  img, _, err := image.Decode(reader)
  if err != nil {
    fmt.Println("Error reading the image, did you fuck it up?")
    panic(err)
  }

  // Compute brightness matrix
  brightness := brightnessMatrix(img)

  // Calculate the width-height ratio of our image
  whR := float64(len(brightness[0]))/float64(len(brightness))

  cwhR := float64(1)
  if f, err := os.OpenFile("/dev/tty", unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NDELAY|unix.O_RDWR, 0666); err == nil {
    var sz *unix.Winsize
    if sz, err = unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ); err == nil {
      // The cell width height ratio 
      cwhR = float64(sz.Xpixel)/float64(sz.Ypixel)
    }
  }

  // The total width height ratio
  twhR := whR * cwhR

  // Average the cells in blocks (specified in "width" by the user)
  averagedBlocks := blockedMatrix(brightness, width, twhR)

  file, err := os.Create(out)
  defer file.Close()
  if err != nil {
    fmt.Println("Error opening file")
    panic(err)
  }
  for y, row := range averagedBlocks{
    prevChar := rune(0)
    count := 0
    for x := range len(row){
      charIndex := float64(averagedBlocks[y][x])/255.0*float64(len(characters)-1)
      invCharIndex := len(characters) - 1 - int(charIndex)
      
      curChar := characters[invCharIndex]
      if prevChar != curChar || x == len(row)-1 {
        // Different character (we add our brainfuck code)
        addBFChars(prevChar, count, file)
        prevChar = curChar
        count = 0
      }
      count++
    }
    addBFNewLine(file)
  }
}

func addBFChars(char rune, count int, file *os.File){
  ascii := int(char)
  addition := ""
  for _ = range ascii{
    addition += "+"
  } 
  for _ = range count{
    addition += "."
  }
  addition +=">\n"
  file.Write([]byte(addition))
}

func addBFNewLine(file *os.File){
  addition := ""
  for _ = range 10 {
    addition += "+"
  }
  addition += ".>\n"
  file.Write([]byte(addition))
}

func brightnessMatrix(img image.Image) [][]int {
  matrix := make([][]int, img.Bounds().Max.Y)
  for i := range matrix {
    matrix[i] = make([]int, img.Bounds().Max.X)
  }
  for y := range img.Bounds().Max.Y {
    for x := range img.Bounds().Max.X {
      pixel := img.At(x, y) 
      r, g, b, _ := pixel.RGBA() // These values need to be divided by 257
      brightness := (r/257 + g/257 + b/257)/3
      matrix[y][x] = int(brightness)
    }
  }
  return matrix
}

func blockedMatrix(matrix [][]int, blocksAmountW int, whR float64) [][]int {
  blockWidth := len(matrix[0]) / blocksAmountW
  blocksAmountH := int(float64(blocksAmountW)/whR)
  blockHeight := len(matrix)/blocksAmountH // first look at the blockheight for a strict emulation of the given aspect ratio
  blocksAmountH = len(matrix)/blockHeight // this will hopefully give us a balance between strict aspect ratio and losslessness

  fmt.Printf("blockWidth: %d, blocksAmountW: %d, width: %d\n", blockWidth, blocksAmountW, len(matrix[0]))
  fmt.Printf("blockHeight: %d, blocksAmountH: %d, height: %d\n", blockHeight, blocksAmountH, len(matrix))
  blockedMatrix := make([][]int, blocksAmountH)
  for i := range len(blockedMatrix) {
    blockedMatrix[i] = make([]int, blocksAmountW)
  }
  for i := range blocksAmountH {
    blockBottom := (i+1)*blockHeight
    blockTop := i*blockHeight 
    for j := range blocksAmountW {
      blockRight := (j+1)*blockWidth
      blockLeft := j*blockWidth
      sum := 0
      for y := blockTop; y < blockBottom; y++ {
          for x := blockLeft; x < blockRight; x++ {
            sum += matrix[y][x]
          }
      }
      average := int(float32(sum) / float32((blockHeight*blockWidth)))
      blockedMatrix[i][j] = average
    }
  }
  return blockedMatrix
}
