package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"strconv"
)

func main() {
  in := os.Args[1:][0]
  out := os.Args[1:][1]
  width, err := strconv.Atoi(os.Args[1:][2])
  if err != nil {
    fmt.Println("The string parameter must be a number")
    panic(err)
  }
  //characters := []rune("$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`'.")
  characters := []rune("Ñ@#W$9876543210?!abc;:+=-,._") 
  reader, err := os.Open(in)
  if err != nil {
    fmt.Println("Error reading your file, did you do it correctly?")
    panic(err)
  }

  img, _, err := image.Decode(reader)
  if err != nil {
    fmt.Println("Error reading the image, did you fuck it up?")
    panic(err)
  }

  brightness := brightnessMatrix(img)
  whR := float32(len(brightness[0]))/float32(len(brightness))
  averagedBlocks := blockedMatrix(brightness, width, whR)

  output := ""
  for y, row := range averagedBlocks{
    prevChar := rune(0)
    count := 0
    for x := range len(row){
      charIndex := float32(averagedBlocks[y][x])/255.0*float32(len(characters)-1)
      invCharIndex := len(characters) - 1 - int(charIndex)
      
      curChar := characters[invCharIndex]
      if prevChar != curChar || x == len(row)-1 {
        // Different character (we add our brainfuck code)
        output = addBFChars(output, prevChar, count)
        prevChar = curChar
        count = 0
      }
      count++
    }
    output = addBFNewLine(output)
  }
  err = os.WriteFile(out, []byte(output), 0644)
  if err != nil {
    fmt.Println("Error writing brainfuck to file")
    panic(err)
  }
}

func addBFChars(input string, char rune, count int) string{
  ascii := int(char)
  for _ = range ascii{
    input += "+"
  } 
  for _ = range count{
    input += "."
  }
  input += ">"
  return input
}

func addBFNewLine(input string) string {
  for _ = range 10 {
    input += "+"
  }
  input += ".>"
  return input
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

func blockedMatrix(matrix [][]int, blocksAmountW int, whR float32) [][]int {
  blockWidth := int(len(matrix[0]) / blocksAmountW)
  blockHeight := int(float32(blockWidth)*whR)
  blocksAmountH := len(matrix)/blockHeight

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
