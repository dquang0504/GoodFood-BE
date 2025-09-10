package integration

var(
	//Seed only 1 account
	SeedAccountOnly = SeedConfig{
		Accounts: &AccountSeed{seedAccount: true,numberOfRecords: 1},
	}

	//Seed account + province/district/ward + address
	SeedMinimalAddress = SeedConfig{
		Accounts:  &AccountSeed{seedAccount: true, numberOfRecords: 1},
		Provinces: true,
		Districts: true,
		Wards:     true,
		Addresses: &AddressSeed{seedAddress: true, numberOfRecords: 1},
	}

	// Seed account + 5 invoices
	SeedAccountWithInvoices = SeedConfig{
		Accounts: &AccountSeed{seedAccount: true, numberOfRecords: 1},
		Invoices: &InvoiceSeed{seedInvoice: true, numberOfRecords: 5},
	}
)