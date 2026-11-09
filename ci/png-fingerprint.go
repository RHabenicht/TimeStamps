//go:build ignore
package main
import (
 "crypto/sha256"
 "fmt"
 "image/png"
 "os"
)
func main(){
 f,err:=os.Open("app/splash_screen.png");if err!=nil{panic(err)};defer f.Close()
 im,err:=png.Decode(f);if err!=nil{panic(err)}
 b:=im.Bounds();h:=sha256.New()
 for y:=b.Min.Y;y<b.Max.Y;y++{for x:=b.Min.X;x<b.Max.X;x++{
  r,g,bb,a:=im.At(x,y).RGBA();h.Write([]byte{byte(r>>8),byte(g>>8),byte(bb>>8),byte(a>>8)})
 }}
 fmt.Printf("Splash pixel dimensions: %dx%d\n",b.Dx(),b.Dy())
 fmt.Printf("Splash RGBA SHA256: %x\n",h.Sum(nil))
 for _,xy:=range [][2]int{{0,0},{100,100},{512,250},{800,450},{1200,450},{1600,800}} {
 r,g,bl,a:=im.At(xy[0],xy[1]).RGBA()
 fmt.Printf("pixel[%d,%d]=%d,%d,%d,%d\n",xy[0],xy[1],r>>8,g>>8,bl>>8,a>>8)
 }
}
