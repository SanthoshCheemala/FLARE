package crypto

import (
	"fmt"
	// "math"
	"github.com/SanthoshCheemala/FLARE.git/pkg/LE"
)


func SetupLEParameters(size int)(*LE.LE,error){
	Q := uint64(180143985094819841)
	qBits := 58
	D := 2048
	N := 4

	var leParams *LE.LE
	var err error

	func (){
		defer func(){
			if r := recover();r != nil {
				err = fmt.Errorf("panic in LE.setup: %v",r)
				fmt.Printf("Recoverd from Panic in LE.setup: %v\n",r)
			}
		} ()
		fmt.Println("Setting up LE with Parameters Q =",Q,"qBits =",qBits,"D =",D,"N =",N)
		leParams = LE.Setup(Q,qBits,D,N)
	}()
	if err != nil{
		return nil,err
	}
	if leParams == nil{
		return nil, fmt.Errorf("failed to initialize the le parameters(nil results)")
	}
	if leParams.R == nil{
		return nil, fmt.Errorf("ring(R) is nil in le parameters")
	}
	// leParams.Layers = int(math.Log10(float64(size)))// minimal depth of the tree for efficiency
	leParams.Layers = 50

	fmt.Println("Successfully initialized the LE parameters: ")
	fmt.Printf(" -Ring Dimension: %d\n",D)
	fmt.Printf(" -Modulus Q: %d\n",Q)
	fmt.Printf(" -Matrix Dimension N: %d\n",N)
	fmt.Printf(" -qBits : %d\n",qBits)
	fmt.Printf(" -Layers: %d\n",leParams.Layers)

	return leParams,nil
}