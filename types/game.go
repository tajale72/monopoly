package types

import (
	"fmt"
	"math/rand"
)

type MonopolyGame interface {
	RollDice()
	Move(position int)
	CheckProperty(property Property) bool
	AddProperty(propertyName string, price int, rent int)
	RemoveBalance(amount int)
	PayRent(rent int)
	PrintProperties()
	PrintDetails()
	PrintAllDetails()
}

func (p *Players) RollDice() {
	fmt.Println(p.Name, "is rolling the dice...")
	// Simulate rolling two dice
	val := 1 + rand.Intn(6) + 1 + rand.Intn(6)
	fmt.Println(p.Name, "rolled a : ", val)
	p.Move(val)
}

func (p *Players) Move(position int) {
	fmt.Println(p.Name, "Moved a : ", position)
	p.Position += position - 1
	if p.Position >= len(AllProperties) {
		p.Position = p.Position % len(AllProperties)
	}

	property := TestDataProperty[p.Position]
	fmt.Println(p.Name, "landed on", property)

	if p.CheckProperty(property) {
		fmt.Println(p.Name, "already owns", property)
		p.PayRent(property.Rent)
		return
	}

	p.AskUserToBuyProperty(property)

	for _, property := range TestDataProperty {
		if p.CheckProperty(property) {
			fmt.Println(p.Name, "already owns", property)
			p.PayRent(property.Rent)
			return
		}
		p.AskUserToBuyProperty(TestDataProperty)

	}

	// for index, property := range AllProperties {
	// 	for name, price := range property {
	// 		if p.Position == index {
	// 			if p.Balance >= price {
	// 				fmt.Println(p.Name, "landed on", name, "with price", price)
	// 				var response string
	// 				fmt.Printf("Would you like to buy this property? (yes/no) , %s: ", p.Name)
	// 				fmt.Scanf("%s", &response)
	// 				fmt.Println("Response:", response)

	// 				if YesChoice[response] {
	// 					p.AddProperty(name, price, 0)
	// 					p.RemoveBalance(price)
	// 					RemoveProperty(index, name)
	// 				} else if response == "no" {
	// 					fmt.Println(p.Name, "decided not to buy", name)
	// 				}
	// 			} else {
	// 				fmt.Println("Not enough balance to buy the property:", name)
	// 				GameEnd = true

	// 			}
	// 			break
	// 		}
	// 	}
	// }
}

func (p *Players) CheckProperty(property Property) bool {
	for _, prop := range p.Properties {
		if prop.PropertyName == property.PropertyName {
			return true
		}
	}
	return false
}

func (p *Players) AskUserToBuyProperty(TestDataProperty map[int]Property) {
	// Check if the property is available for purchase
	for index, property := range TestDataProperty {
		if p.Position != index {
			continue
		}
		if property.Owner == "" && p.Balance >= property.Rent {
			fmt.Println(p.Name, "landed on", property.PropertyName, "with price", property.Price)

			var response string
			fmt.Printf("Would you like to buy this property? %s (yes/no) , %s: ", property.PropertyName, p.Name)
			fmt.Scanf("%s", &response)

			if YesChoice[response] {
				p.AddProperty(property.PropertyName, property.Price, property.Rent)
				p.RemoveBalance(property.Price)
				RemoveProperty(p.Position, property.PropertyName)
			}
			if NoChoice[response] {
				fmt.Println(p.Name, "decided not to buy", property.PropertyName)
			}
		} else {
			fmt.Println("Not enough balance to buy the property:", property.PropertyName)
		}

	}

}

func (p *Players) AddProperty(propertyName string, price int, rent int) {
	propertyDetials := Property{
		PropertyName: propertyName,
		Price:        price, // Price will be set later when buying
		Rent:         rent,  // Rent will be set later when buying
		Owner:        p.Name,
	}
	fmt.Printf("Adding property  %s to %s \n", propertyName, p.Name)
	p.Properties = append(p.Properties, propertyDetials)
}

func (p *Players) RemoveBalance(amount int) {
	fmt.Printf("Removing balance  %d from %s \n", amount, p.Name)
	p.Balance -= amount
	fmt.Printf("Current balance for user %s from %d \n", p.Name, p.Balance)

}

func RemoveProperty(indexToRemove int, propertyName string) {
	fmt.Println("Removing property:", propertyName)
	AllProperties = append(AllProperties[:indexToRemove], AllProperties[indexToRemove+1:]...)
}

func (p *Players) PayRent(rent int) {
	fmt.Println(p.Name, "paying rent to :", rent)
	p.Balance -= rent
	if p.Balance < 0 {
		fmt.Println(p.Name, "has run out of balance. Game Over!")
		GameEnd = true
	}
	fmt.Println("Current balance after paying rent:", p.Balance)
}

func (p *Players) PrintProperties() {
	fmt.Println(p.Name, "Properties:")
	for _, property := range p.Properties {
		fmt.Printf("Property: %s, Price: %d, Rent: %d, Owner: %s\n", property.PropertyName, property.Price, property.Rent, property.Owner)
	}
}

func (p *Players) PrintDetails() {
	fmt.Printf("Player: %s, Balance: %d, Position: %d\n", p.Name, p.Balance, p.Position)
	p.PrintProperties()
}
func (p *Players) PrintAllDetails() {
	fmt.Println("Player Details:")
	fmt.Printf("Name: %s, Balance: %d, Position: %d\n", p.Name, p.Balance, p.Position)
	p.PrintProperties()
}
