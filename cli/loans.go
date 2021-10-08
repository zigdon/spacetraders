package cli

import (
	"fmt"
	"log"
	"time"

	"github.com/zigdon/spacetraders"
)

func init() {
	for _, c := range []cmd{
		{
			Section: "Loans",
			Name:    "AvailableLoans",
			Usage:   "AvailableLoans",
			Help:    "Display currently available loans",
			Do:      doLoans,
			Aliases: []string{"lsLoans"},
		},
		{
			Section: "Loans",
			Name:    "TakeLoan",
			Usage:   "TakeLoan <type>",
			Help:    "Take out one of the available loans",
			Do:      doTakeLoan,
			MinArgs: 1,
			MaxArgs: 1,
		},
		{
			Section: "Loans",
			Name:    "MyLoans",
			Usage:   "MyLoans",
			Help:    "List outstanding loans",
			Do:      doMyLoans,
		},
		{
			Section:    "Loans",
			Name:       "PayLoan",
			Usage:      "PayLoan <loanID>",
			Validators: []string{"loans"},
			Help:       "Pay an outstanding loan",
			Do:         doPayLoan,
			MinArgs: 1,
			MaxArgs: 1,
		},
	} {
		if err := Register(c); err != nil {
			log.Fatalf("Can't register %q: %v", c.Name, err)
		}
	}
}

func doLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.AvailableLoans()
	if err != nil {
		return fmt.Errorf("error getting loans: %v", err)
	}

	for _, l := range loans {
		Out("amt: %d, needs collateral: %v, rate: %d, term (days): %d, type: %s",
			l.Amount, l.CollateralRequired, l.Rate, l.TermInDays, l.Type)
	}

	return nil
}

func doTakeLoan(c *spacetraders.Client, args []string) error {
	loan, err := c.TakeLoan(args[0])
	if err != nil {
		return fmt.Errorf("error taking out loan: %v", err)
	}

	Out("Loan taken, %s (%s), due: %s (in %s)",
		loan.ShortID, loan.ID, loan.Due.Local(), loan.Due.Sub(time.Now()).Truncate(time.Second))

	return nil
}

func doMyLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.MyLoans()
	if err != nil {
		return fmt.Errorf("error querying loans: %v", err)
	}

	for _, l := range loans {
		Out(l.String())
	}

	return nil
}

func doPayLoan(c *spacetraders.Client, args []string) error {
    if len(args) == 0 {
	  return fmt.Errorf("missing args for loan")
	}
	err := c.PayLoan(args[0])
	if err != nil {
		return fmt.Errorf("error paying loan off: %v", err)
	}

	Out("Loan paid. Current loans:")
	return doMyLoans(c, args)
}
