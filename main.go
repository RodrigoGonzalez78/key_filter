package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bendahl/uinput"
	evdev "github.com/gvalkov/golang-evdev"
)

var correcciones = map[uint16]uint16{
	evdev.KEY_R: evdev.KEY_Y, // ry -> r
	evdev.KEY_T: evdev.KEY_Y, // ty -> t
	evdev.KEY_F: evdev.KEY_H, // fh -> f
	evdev.KEY_G: evdev.KEY_H, // gh -> g
	evdev.KEY_V: evdev.KEY_N, // vn -> v
	evdev.KEY_B: evdev.KEY_N, // bn -> b
	evdev.KEY_4: evdev.KEY_6, // 46 -> 4
	evdev.KEY_5: evdev.KEY_6, // 56 -> 5
}

func esErrorMecanico(teclaActual uint16, teclaAnterior uint16, tiempoAnterior time.Time, limite time.Duration) bool {
	// Si pasó más tiempo que el límite (ej. 30ms), es porque tú escribiste la letra a propósito.
	if time.Since(tiempoAnterior) > limite {
		return false
	}

	// Si el tiempo es muy corto, revisamos el diccionario
	teclaFantasmaEsperada, existe := correcciones[teclaAnterior]

	// Si la combinación está en el diccionario, la bloqueamos
	if existe && teclaFantasmaEsperada == teclaActual {
		return true
	}

	return false
}

func main() {
	// IMPORTANTE: Aquí pondremos la ruta de tu teclado físico
	devicePath := "/dev/input/event2"

	device, err := evdev.Open(devicePath)
	if err != nil {
		log.Fatalf("Error abriendo el teclado: %v", err)
	}
	defer device.File.Close()

	device.Grab()
	defer device.Release()

	keyboard, err := uinput.CreateKeyboard("/dev/uinput", []byte("Filtro Go Seguro"))
	if err != nil {
		log.Fatalf("Error creando teclado virtual: %v", err)
	}
	defer keyboard.Close()

	fmt.Println("Filtro genérico iniciado. Esperando tecleos...")

	var ultimaTecla uint16
	var ultimoTiempo time.Time

	// Milisegundos de tolerancia (ajústalo si escribes muy rápido o muy lento)
	umbral := 45 * time.Millisecond

	for {
		event, err := device.ReadOne()
		if err != nil {
			log.Fatal(err)
		}

		if event.Type == evdev.EV_KEY {
			esPresion := event.Value == 1
			esLiberacion := event.Value == 0

			if esPresion {
				// 3. LLAMAMOS A LA FUNCIÓN GENÉRICA
				if esErrorMecanico(event.Code, ultimaTecla, ultimoTiempo, umbral) {
					fmt.Printf("Bloqueo exitoso: Tecla %d ignorada.\n", event.Code)
					continue // Salta al inicio del bucle, ignorando esta pulsación
				}

				// Actualizamos el estado para la próxima vuelta
				ultimaTecla = event.Code
				ultimoTiempo = time.Now()
			}

			if esPresion {
				keyboard.KeyDown(int(event.Code))
			} else if esLiberacion {
				keyboard.KeyUp(int(event.Code))
			}
		}
	}
}
