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

	// Seed account + 12 invoices
	SeedAccountWithInvoicesNoDetail = SeedConfig{
		Accounts: &AccountSeed{seedAccount: true, numberOfRecords: 1},
		InvoiceStatuses: true,
		Invoices: &InvoiceSeed{seedInvoice: true, numberOfRecords: 12},
	}

	//Seed invoice no detail
	SeedHappyPathInvoice = SeedConfig{
		Accounts: &AccountSeed{seedAccount: true,numberOfRecords: 1},
		ProductTypes: &ProductTypeSeed{seedProductType: true,numberOfRecords: 6},
		Products: &ProductSeed{seedProduct: true,numberOfRecords: 6},
		InvoiceStatuses: true,
		Invoices: &InvoiceSeed{seedInvoice: true,numberOfRecords: 6},
		InvoiceDetails: &InvoiceDetailSeed{seedInvoiceDetail: true,numberOfRecords: 6},
	}
)