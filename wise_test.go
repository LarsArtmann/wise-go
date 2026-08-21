package wise_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larsartmann/wise-go"
	"github.com/larsartmann/wise-go/internal/raw"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWiseClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Wise SDK Suite")
}

func stdBalance(
	id int64,
	currency, name string,
	visible bool,
	investmentState string,
	amount, reserved float64,
	created string,
) raw.Balance {
	return raw.Balance{
		ID: id, Currency: currency, Type: "STANDARD", Name: name,
		InvestmentState: investmentState, Visible: visible,
		Amount:         raw.BalanceAmount{Value: amount, Currency: currency},
		ReservedAmount: raw.BalanceAmount{Value: reserved, Currency: currency},
		CreationTime:   created,
	}
}

func testTx(id, txType string, amount float64) raw.StatementTransaction {
	return raw.StatementTransaction{
		TransactionID: id, Date: "2023-01-15 14:30:00",
		Amount:         raw.BalanceAmount{Value: amount, Currency: "EUR"},
		TotalFees:      raw.BalanceAmount{Value: 0, Currency: "EUR"},
		RunningBalance: raw.BalanceAmount{Value: 0, Currency: "EUR"},
		Details:        raw.TransactionDetails{Type: txType},
	}
}

var unauthorizedHandler = func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"code":"UNAUTHORIZED","message":"Invalid API key"}]}`))
}

func expectListProfilesError(client *wise.Client, substr string) {
	_, err := client.ListProfiles(context.Background())
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring(substr))
}

func testProfiles(id int64, firstName, lastName, email string) []raw.Profile {
	return []raw.Profile{
		{
			ID:        id,
			Type:      "PERSONAL",
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			CreatedAt: "2023-01-01T00:00:00Z",
		},
	}
}

func personalProfile(id int64, firstName, lastName, email, createdAt string) raw.Profile {
	return raw.Profile{
		ID: id, Type: "PERSONAL",
		FirstName: firstName, LastName: lastName,
		Email: email, CreatedAt: createdAt,
	}
}

func expectProfileID(profiles []wise.Profile, idx int, expectedID int64) {
	Expect(profiles[idx].ID.Get()).To(Equal(expectedID))
}

func expectBalanceAmountCents(
	balances []wise.Balance, idx int, expectedAmount, expectedReserved int64,
) {
	Expect(balances[idx].Amount.Cents).To(Equal(expectedAmount))
	Expect(balances[idx].Reserved.Cents).To(Equal(expectedReserved))
}

func expectTransactionQueryParams(r *http.Request, currency, intervalStart, intervalEnd string) {
	params := map[string]string{
		"currency":      currency,
		"intervalStart": intervalStart,
		"intervalEnd":   intervalEnd,
	}
	for name, expected := range params {
		Expect(r.URL.Query().Get(name)).To(Equal(expected))
	}
}

func listBalances(ctx context.Context, client *wise.Client) ([]wise.Balance, error) {
	//nolint:wrapcheck // transparent pass-through; tests assert on the original error.
	return client.ListBalances(ctx, wise.NewProfileID(12345))
}

func getBalance(
	ctx context.Context, client *wise.Client, balanceID int64,
) (*wise.Balance, error) {
	//nolint:wrapcheck // transparent pass-through; tests assert on the original error.
	return client.GetBalance(ctx, wise.NewProfileID(12345), wise.NewBalanceID(balanceID))
}

var _ = Describe("Wise Client", func() {
	var (
		server           *httptest.Server
		mux              *http.ServeMux
		client           *wise.Client
		defaultListTxReq wise.ListTransactionsRequest
	)

	BeforeEach(func() {
		mux = http.NewServeMux()
		server = httptest.NewServer(mux)
		client = wise.New("test-api-key", wise.WithBaseURL(server.URL))
		defaultListTxReq = wise.ListTransactionsRequest{
			ProfileID: wise.NewProfileID(12345),
			BalanceID: wise.NewBalanceID(100),
			Currency:  wise.Currency("EUR"),
			From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("New", func() {
		It("should create client with default options", func() {
			c := wise.New("key")
			Expect(c).ToNot(BeNil())
		})

		It("should create client with sandbox option", func() {
			c := wise.New("key", wise.WithSandbox())
			Expect(c).ToNot(BeNil())
		})

		It("should create client with custom timeout", func() {
			c := wise.New("key", wise.WithTimeout(10*time.Second))
			Expect(c).ToNot(BeNil())
		})

		It("should create client with custom retry", func() {
			c := wise.New("key", wise.WithRetry(5, time.Second, 30*time.Second))
			Expect(c).ToNot(BeNil())
		})
	})

	Describe("Authenticate", func() {
		Context("with valid credentials", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Header.Get("Authorization")).To(Equal("Bearer test-api-key"))
					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(
						w,
						testProfiles(12345, "John", "Doe", "john@example.com"),
					)
				})
			})

			It("should not return error", func() {
				err := client.Authenticate(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with invalid credentials", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", unauthorizedHandler)
			})

			It("should return an AuthError", func() {
				err := client.Authenticate(context.Background())
				Expect(err).To(HaveOccurred())

				_, ok := errors.AsType[*wise.AuthError](err)
				Expect(ok).To(BeTrue())
			})
		})
	})

	Describe("Health", func() {
		It("should delegate to Authenticate", func() {
			mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})

			err := client.Health(context.Background())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListProfiles", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
					profiles := []raw.Profile{
						personalProfile(
							12345, "John", "Doe", "john@example.com", "2023-01-15T10:30:00Z",
						),
						{
							ID: 67890, Type: "BUSINESS",
							BusinessName: "Acme Corp",
							Email:        "billing@acme.com", CreatedAt: "2023-02-20T14:00:00Z",
						},
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, profiles)
				})
			})

			It("should return mapped profiles", func() {
				profiles, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(profiles).To(HaveLen(2))
			})

			It("should map personal profile correctly", func() {
				profiles, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())

				expectProfileID(profiles, 0, 12345)
				Expect(profiles[0].Name).To(Equal("John Doe"))
				Expect(profiles[0].Type).To(Equal(wise.ProfileTypePersonal))
				Expect(profiles[0].Email).To(Equal("john@example.com"))
			})

			It("should map business profile correctly", func() {
				profiles, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())

				expectProfileID(profiles, 1, 67890)
				Expect(profiles[1].Name).To(Equal("Acme Corp"))
				Expect(profiles[1].Type).To(Equal(wise.ProfileTypeBusiness))
			})

			It("should parse created_at timestamp", func() {
				profiles, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())

				Expect(profiles[0].CreatedAt.Year()).To(Equal(2023))
				Expect(profiles[0].CreatedAt.Month()).To(Equal(time.January))
			})
		})

		Context("with API error", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", unauthorizedHandler)
			})

			It("should return error with API error details", func() {
				expectListProfilesError(client, "UNAUTHORIZED")
			})
		})

		Context("with zoneless createdAt (live /v2/profiles format)", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
					profiles := []raw.Profile{
						// Exact createdAt value observed from the live Wise API
						// on 2026-08-18: no zone designator.
						personalProfile(14757634, "Lars", "Artmann", "lars@example.com", "2020-05-27T10:27:22"),
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, profiles)
				})
			})

			It("should parse createdAt as UTC and return the profile", func() {
				profiles, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(profiles).To(HaveLen(1))
				Expect(profiles[0].CreatedAt).To(Equal(time.Date(2020, time.May, 27, 10, 27, 22, 0, time.UTC)))
			})
		})

		Context("with unknown profile type", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
					profiles := []raw.Profile{
						{ID: 1, Type: "UNKNOWN_TYPE", CreatedAt: "2023-01-15T10:30:00Z"},
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, profiles)
				})
			})

			It("should return an error", func() {
				expectListProfilesError(client, "unknown profile type")
			})
		})
	})

	Describe("ListBalances", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v4/profiles/12345/balances",
					func(w http.ResponseWriter, r *http.Request) {
						Expect(r.URL.Query().Get("types")).To(Equal("STANDARD,SAVINGS"))

						balances := []raw.Balance{
							stdBalance(
								100,
								"EUR",
								"Main Account",
								true,
								"NOT_INVESTED",
								1234.56,
								0.0,
								"2023-01-01T00:00:00Z",
							),
							stdBalance(
								200,
								"USD",
								"USD Account",
								true,
								"NOT_INVESTED",
								500.00,
								50.00,
								"2023-01-02T00:00:00Z",
							),
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, balances)
					},
				)
			})

			It("should return visible balances", func() {
				balances, err := listBalances(context.Background(), client)
				Expect(err).ToNot(HaveOccurred())
				Expect(balances).To(HaveLen(2))
			})

			It("should map balance fields correctly", func() {
				balances, err := listBalances(context.Background(), client)
				Expect(err).ToNot(HaveOccurred())

				Expect(balances[0].ID.Get()).To(Equal(int64(100)))
				Expect(balances[0].Currency).To(Equal(wise.Currency("EUR")))
				Expect(balances[0].Type).To(Equal(wise.BalanceTypeStandard))
				Expect(balances[0].Name).To(Equal("Main Account"))
			})

			It("should convert amounts to cents", func() {
				balances, err := listBalances(context.Background(), client)
				Expect(err).ToNot(HaveOccurred())

				expectBalanceAmountCents(balances, 0, 123456, 0)
				expectBalanceAmountCents(balances, 1, 50000, 5000)
			})
		})

		Context("with invisible balances", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v4/profiles/12345/balances",
					func(w http.ResponseWriter, _ *http.Request) {
						balances := []raw.Balance{
							stdBalance(
								100,
								"EUR",
								"Visible",
								true,
								"NOT_INVESTED",
								100.00,
								0,
								"2023-01-01T00:00:00Z",
							),
							stdBalance(
								200,
								"EUR",
								"Invisible",
								false,
								"NOT_INVESTED",
								200.00,
								0,
								"2023-01-01T00:00:00Z",
							),
							stdBalance(
								300,
								"EUR",
								"Investment",
								true,
								"INVESTED",
								300.00,
								0,
								"2023-01-01T00:00:00Z",
							),
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, balances)
					},
				)
			})

			It("should filter out invisible and investment balances", func() {
				balances, err := listBalances(context.Background(), client)
				Expect(err).ToNot(HaveOccurred())
				Expect(balances).To(HaveLen(1))
				Expect(balances[0].Name).To(Equal("Visible"))
			})
		})

		Context("with API error", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v4/profiles/12345/balances",
					func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
						_, _ = w.Write(
							[]byte(
								`{"errors":[{"code":"SERVER_ERROR","message":"Internal server error"}]}`,
							),
						)
					},
				)
			})

			It("should return error", func() {
				_, err := listBalances(context.Background(), client)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetBalance", func() {
		BeforeEach(func() {
			mux.HandleFunc(
				"/v4/profiles/12345/balances",
				func(w http.ResponseWriter, _ *http.Request) {
					balances := []raw.Balance{
						stdBalance(
							100,
							"EUR",
							"EUR Account",
							true,
							"NOT_INVESTED",
							1000.00,
							0,
							"2023-01-01T00:00:00Z",
						),
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, balances)
				},
			)
		})

		It("should return the specific balance", func() {
			balance, err := getBalance(context.Background(), client, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(balance.ID.Get()).To(Equal(int64(100)))
			Expect(balance.Amount.Cents).To(Equal(int64(100000)))
		})

		It("should return error for non-existent balance", func() {
			_, err := getBalance(context.Background(), client, 999)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Describe("ListTransactions", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, r *http.Request) {
						expectTransactionQueryParams(
							r, "EUR", "2023-01-01T00:00:00Z", "2023-01-31T23:59:59Z",
						)

						response := raw.StatementResponse{
							Transactions: []raw.StatementTransaction{
								{
									TransactionID: "tx-001",
									Date:          "2023-01-15 14:30:00",
									Amount: raw.BalanceAmount{
										Value:    -50.00,
										Currency: "EUR",
									},
									TotalFees: raw.BalanceAmount{
										Value:    0.50,
										Currency: "EUR",
									},
									RunningBalance: raw.BalanceAmount{
										Value:    950.50,
										Currency: "EUR",
									},
									ReferenceNumber: "REF-001",
									Details: raw.TransactionDetails{
										Type:         "CARD_PAYMENT",
										Description:  "Coffee Shop",
										Category:     "food",
										MerchantName: "Starbucks",
									},
								},
								{
									TransactionID: "tx-002",
									Date:          "2023-01-20 10:00:00",
									Amount: raw.BalanceAmount{
										Value:    1000.00,
										Currency: "EUR",
									},
									TotalFees:       raw.BalanceAmount{Value: 0, Currency: "EUR"},
									RunningBalance:  raw.BalanceAmount{Value: 1950.50, Currency: "EUR"},
									ReferenceNumber: "REF-002",
									Details: raw.TransactionDetails{
										Type:        "TRANSFER",
										Description: "Salary Payment",
										Category:    "income",
									},
								},
							},
							EndOfStatementBalance: raw.BalanceAmount{
								Value:    950.50,
								Currency: "EUR",
							},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, response)
					},
				)
			})

			It("should return transactions", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions).To(HaveLen(2))
			})

			It("should map transaction ID correctly", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].ID.Get()).To(Equal("tx-001"))
			})

			It("should classify card payment as card type", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].Type).To(Equal(wise.TransactionTypeCard))
			})

			It("should classify transfer as transfer type", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[1].Type).To(Equal(wise.TransactionTypeTransfer))
			})

			It("should convert amounts to cents correctly", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())

				// First transaction: -50.00 → totalCents=-5000, amountCents=5000 (abs)
				Expect(resp.Transactions[0].Total.Cents).To(Equal(int64(-5000)))
				Expect(resp.Transactions[0].Amount.Cents).To(Equal(int64(5000)))
				Expect(resp.Transactions[0].Fees.Cents).To(Equal(int64(50)))
			})

			It("should parse date correctly", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].Date.Year()).To(Equal(2023))
				Expect(resp.Transactions[0].Date.Month()).To(Equal(time.January))
				Expect(resp.Transactions[0].Date.Day()).To(Equal(15))
			})

			It("should map description and merchant", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].Description).To(Equal("Coffee Shop"))
				Expect(resp.Transactions[0].MerchantName).To(Equal("Starbucks"))
				Expect(resp.Transactions[0].Category).To(Equal("food"))
				Expect(resp.Transactions[0].Reference).To(Equal("REF-001"))
			})

			It("should surface end-of-statement balance", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.EndOfStatementBalance.Cents).To(Equal(int64(95050)))
				Expect(resp.EndOfStatementBalance.Currency).To(Equal(wise.Currency("EUR")))
			})
		})

		Context("with transaction type edge cases", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						response := raw.StatementResponse{
							Transactions: []raw.StatementTransaction{
								testTx("tx-refund", "CARD_REFUND", 25.00),
								testTx("tx-exchange", "CONVERSION", -100.00),
								testTx("tx-fee", "FEE", -0.50),
								testTx("tx-payment", "PAYMENT", -200.00),
							},
							EndOfStatementBalance: raw.BalanceAmount{Value: 0, Currency: "EUR"},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, response)
					},
				)
			})

			It("should classify CARD_REFUND with positive amount as refund", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].Type).To(Equal(wise.TransactionTypeRefund))
			})

			It("should classify CONVERSION as exchange", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[1].Type).To(Equal(wise.TransactionTypeExchange))
			})

			It("should classify FEE as fee", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[2].Type).To(Equal(wise.TransactionTypeFee))
			})

			It("should classify PAYMENT as payment", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[3].Type).To(Equal(wise.TransactionTypePayment))
			})
		})

		Context("with API error", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write(
							[]byte(
								`{"errors":[{"code":"NOT_FOUND","message":"Balance not found"}]}`,
							),
						)
					},
				)
			})

			It("should return NotFoundError", func() {
				_, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).To(HaveOccurred())

				_, ok := errors.AsType[*wise.NotFoundError](err)
				Expect(ok).To(BeTrue())
			})
		})

		Context("with running balance and exchange details", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						response := raw.StatementResponse{
							Transactions: []raw.StatementTransaction{
								{
									TransactionID: "tx-exch",
									Date:          "2023-01-15 14:30:00",
									Amount:        raw.BalanceAmount{Value: -100.00, Currency: "EUR"},
									TotalFees:     raw.BalanceAmount{Value: 0, Currency: "EUR"},
									RunningBalance: raw.BalanceAmount{
										Value: 5000.00, Currency: "EUR",
									},
									ReferenceNumber: "REF-EXCH",
									Details:         raw.TransactionDetails{Type: "CONVERSION"},
									ExchangeDetails: &raw.ExchangeDetails{
										FromAmount:   raw.BalanceAmount{Value: 100.00, Currency: "EUR"},
										ToAmount:     raw.BalanceAmount{Value: 108.50, Currency: "USD"},
										Rate:         1.085,
										FromCurrency: "EUR",
										ToCurrency:   "USD",
									},
								},
							},
							EndOfStatementBalance: raw.BalanceAmount{Value: 5000, Currency: "EUR"},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, response)
					},
				)
			})

			It("should map running balance", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].RunningBalance.Cents).To(Equal(int64(500000)))
				Expect(resp.Transactions[0].RunningBalance.Currency).To(Equal(wise.Currency("EUR")))
			})

			It("should map exchange details", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())

				exch := resp.Transactions[0].Exchange
				Expect(exch).ToNot(BeNil())
				Expect(exch.From.Cents).To(Equal(int64(10000)))
				Expect(exch.From.Currency).To(Equal(wise.Currency("EUR")))
				Expect(exch.To.Cents).To(Equal(int64(10850)))
				Expect(exch.To.Currency).To(Equal(wise.Currency("USD")))
				Expect(exch.Rate).To(Equal(1.085))
			})
		})

		Context("with a cross-currency transaction", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						response := raw.StatementResponse{
							Transactions: []raw.StatementTransaction{
								{
									TransactionID: "tx-usd",
									Date:          "2023-01-15 14:30:00",
									Amount:        raw.BalanceAmount{Value: -50.00, Currency: "USD"},
									TotalFees:     raw.BalanceAmount{Value: 0, Currency: "USD"},
									RunningBalance: raw.BalanceAmount{
										Value: 200.00, Currency: "USD",
									},
									ReferenceNumber: "REF-USD",
									Details:         raw.TransactionDetails{Type: "CARD_PAYMENT"},
								},
							},
							EndOfStatementBalance: raw.BalanceAmount{Value: 200, Currency: "USD"},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, response)
					},
				)
			})

			It("should use the transaction currency, not the request currency", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())

				tx := resp.Transactions[0]
				Expect(tx.Amount.Currency).To(Equal(wise.Currency("USD")))
				Expect(tx.Total.Currency).To(Equal(wise.Currency("USD")))
				Expect(tx.RunningBalance.Currency).To(Equal(wise.Currency("USD")))
			})
		})

		Context("with a zero end-of-statement balance", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						response := raw.StatementResponse{
							Transactions:          []raw.StatementTransaction{},
							EndOfStatementBalance: raw.BalanceAmount{Value: 0, Currency: "EUR"},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, response)
					},
				)
			})

			It("should return a zero Money value without error", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions).To(BeEmpty())
				Expect(resp.EndOfStatementBalance.Cents).To(Equal(int64(0)))
				Expect(resp.EndOfStatementBalance.Currency).To(Equal(wise.Currency("EUR")))
			})
		})

		Context("with a type filter on the request", func() {
			var capturedRequest *http.Request

			BeforeEach(func() {
				capturedRequest = nil

				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, r *http.Request) {
						capturedRequest = r
						response := raw.StatementResponse{
							Transactions: []raw.StatementTransaction{
								testTx("tx-filtered", "CARD_PAYMENT", -10.00),
							},
							EndOfStatementBalance: raw.BalanceAmount{Value: 0, Currency: "EUR"},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, response)
					},
				)
			})

			It("should forward the type filter in the query string", func() {
				req := defaultListTxReq
				req.Type = wise.DetailTypeCardPayment

				_, err := client.ListTransactions(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(capturedRequest).ToNot(BeNil())
				Expect(capturedRequest.URL.Query().Get("type")).To(Equal("CARD_PAYMENT"))
			})

			It("should omit the type filter when unset", func() {
				_, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(capturedRequest).ToNot(BeNil())
				_, hasType := capturedRequest.URL.Query()["type"]
				Expect(hasType).To(BeFalse())
			})
		})

		Context("with local-zone timestamps", func() {
			// Regression: Wise rejects zone offsets other than "Z" with
			// HTTP 422 wrong.date.format. Callers pass local time.Time
			// values; the client must normalize intervalStart/End to UTC.
			var cest *time.Location

			BeforeEach(func() {
				cest = time.FixedZone("CEST", 2*60*60)

				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, r *http.Request) {
						expectTransactionQueryParams(
							r, "EUR", "2023-01-01T00:00:00Z", "2023-01-31T23:59:59Z",
						)

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, raw.StatementResponse{
							EndOfStatementBalance: raw.BalanceAmount{Value: 0, Currency: "EUR"},
						})
					})
			})

			It("normalizes intervalStart/End to UTC Z format", func() {
				req := defaultListTxReq
				req.From = time.Date(2023, 1, 1, 2, 0, 0, 0, cest) // 2023-01-01T00:00:00Z
				req.To = time.Date(2023, 2, 1, 1, 59, 59, 0, cest) // 2023-01-31T23:59:59Z
				resp, err := client.ListTransactions(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions).To(BeEmpty())
			})
		})

		Context("with invalid request", func() {
			It("should reject missing currency", func() {
				req := defaultListTxReq
				req.Currency = ""
				_, err := client.ListTransactions(context.Background(), req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("currency is required"))
			})

			It("should reject From after To", func() {
				req := defaultListTxReq
				req.From = time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
				req.To = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				_, err := client.ListTransactions(context.Background(), req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("intervalStart must not be after intervalEnd"))
			})
		})
	})

	Describe("ListTransfers", func() {
		transferFixture := func(id int64, status string, created string) raw.Transfer {
			sourceAccount := int64(5678901)

			return raw.Transfer{
				ID: id, User: 4342275, TargetAccount: 8692237,
				SourceAccount: &sourceAccount,
				Status:        status, Rate: 0.89, Created: created,
				Details:         raw.TransferDetails{Reference: "Rent November"},
				HasActiveIssues: false,
				SourceCurrency:  "EUR", SourceValue: 168.54,
				TargetCurrency: "GBP", TargetValue: 150.0,
				CustomerTransactionID: "54a6bc09-cef9-49a8-9041-f1f0c654cd88",
			}
		}

		Context("with a single page", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/transfers", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Query().Get("profile")).To(Equal("12345"))
					Expect(r.URL.Query().Get("createdDateStart")).To(Equal("2023-01-01T00:00:00Z"))
					Expect(r.URL.Query().Get("createdDateEnd")).To(Equal("2023-12-31T23:59:59Z"))
					Expect(r.URL.Query().Get("status")).To(Equal("delivered,cancelled"))
					Expect(r.URL.Query().Get("limit")).To(Equal("100"))

					response := []raw.Transfer{
						transferFixture(16521632, "delivered", "2023-11-24 10:47:49"),
						func() raw.Transfer { // Wise omits sourceAccount on some transfers.
							withoutSource := transferFixture(16521633, "cancelled", "2023-12-02T08:15:00Z")
							withoutSource.SourceAccount = nil

							return withoutSource
						}(),
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, response)
				})
			})

			It("maps transfer fields", func() {
				transfers, err := client.ListTransfers(context.Background(), wise.ListTransfersRequest{
					ProfileID: wise.NewProfileID(12345),
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
					Status:    []wise.TransferStatus{wise.TransferStatusDelivered, wise.TransferStatusCancelled},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(transfers).To(HaveLen(2))

				first := transfers[0]
				Expect(first.ID.Get()).To(Equal(int64(16521632)))
				Expect(first.RecipientID.Get()).To(Equal(int64(8692237)))
				Expect(first.Status).To(Equal(wise.TransferStatusDelivered))
				Expect(first.Rate).To(Equal(0.89))
				Expect(first.Source.Cents).To(Equal(int64(16854)))
				Expect(first.Source.Currency).To(Equal(wise.Currency("EUR")))
				Expect(first.Target.Cents).To(Equal(int64(15000)))
				Expect(first.Target.Currency).To(Equal(wise.Currency("GBP")))
				Expect(first.Created).To(Equal(
					time.Date(2023, 11, 24, 10, 47, 49, 0, time.UTC),
				))
				Expect(first.Reference).To(Equal("Rent November"))
				Expect(first.CustomerTransactionID).To(Equal("54a6bc09-cef9-49a8-9041-f1f0c654cd88"))
				Expect(first.HasActiveIssues).To(BeFalse())
			})

			It("maps the debited balance and preserves an omitted sourceAccount as nil", func() {
				transfers, err := client.ListTransfers(context.Background(), wise.ListTransfersRequest{
					ProfileID: wise.NewProfileID(12345),
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
					Status:    []wise.TransferStatus{wise.TransferStatusDelivered, wise.TransferStatusCancelled},
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(transfers[0].SourceAccount).ToNot(BeNil())
				Expect(transfers[0].SourceAccount.Get()).To(Equal(int64(5678901)))
				Expect(transfers[1].SourceAccount).To(BeNil())
			})

			It("parses both space-separated and RFC3339 created timestamps", func() {
				transfers, err := client.ListTransfers(context.Background(), wise.ListTransfersRequest{
					ProfileID: wise.NewProfileID(12345),
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
					Status:    []wise.TransferStatus{wise.TransferStatusDelivered, wise.TransferStatusCancelled},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(transfers[0].Created.Format(time.RFC3339)).To(Equal("2023-11-24T10:47:49Z"))
				Expect(transfers[1].Created.Format(time.RFC3339)).To(Equal("2023-12-02T08:15:00Z"))
			})
		})

		Context("with local-zone timestamps", func() {
			// Regression: Wise rejects zone offsets other than "Z" with
			// HTTP 422 wrong.date.format (seen live 2026-08-19..21 with
			// createdDateEnd=…+02:00). Callers pass local time.Time values
			// (e.g. time.Now()); the client must normalize to UTC on the wire.
			var cest *time.Location

			BeforeEach(func() {
				cest = time.FixedZone("CEST", 2*60*60)

				mux.HandleFunc("/v1/transfers", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Query().Get("createdDateStart")).To(Equal("2023-01-01T00:00:00Z"))
					Expect(r.URL.Query().Get("createdDateEnd")).To(Equal("2023-12-31T23:59:59Z"))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, []raw.Transfer{})
				})
			})

			It("normalizes createdDateStart/End to UTC Z format", func() {
				transfers, err := client.ListTransfers(context.Background(), wise.ListTransfersRequest{
					ProfileID: wise.NewProfileID(12345),
					From:      time.Date(2023, 1, 1, 2, 0, 0, 0, cest),   // 2023-01-01T00:00:00Z
					To:        time.Date(2024, 1, 1, 1, 59, 59, 0, cest), // 2023-12-31T23:59:59Z
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(transfers).To(BeEmpty())
			})
		})

		Context("across multiple pages", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/transfers", func(w http.ResponseWriter, r *http.Request) {
					offset := r.URL.Query().Get("offset")

					var response []raw.Transfer

					if offset == "0" {
						response = make([]raw.Transfer, 0, 100)
						for i := range 100 {
							response = append(
								response,
								transferFixture(int64(1000+i), "delivered", "2023-11-24 10:47:49"),
							)
						}
					} else {
						response = []raw.Transfer{transferFixture(2000, "delivered", "2023-12-01 09:00:00")}
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, response)
				})
			})

			It("fetches until a short page", func() {
				transfers, err := client.ListTransfers(context.Background(), wise.ListTransfersRequest{
					ProfileID: wise.NewProfileID(12345),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(transfers).To(HaveLen(101))
				Expect(transfers[100].ID.Get()).To(Equal(int64(2000)))
			})
		})

		Context("with an unparseable created timestamp", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/transfers", func(w http.ResponseWriter, _ *http.Request) {
					response := []raw.Transfer{transferFixture(1, "delivered", "not-a-timestamp")}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, response)
				})
			})

			It("returns a corruption error naming the transfer", func() {
				_, err := client.ListTransfers(context.Background(), wise.ListTransfersRequest{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("map transfer 1"))
			})
		})
	})

	Describe("GetProfile", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles/12345", func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, personalProfile(
						12345, "John", "Doe", "john@example.com", "2023-01-15T10:30:00Z",
					))
				})
			})

			It("should return the mapped profile", func() {
				profile, err := client.GetProfile(context.Background(), wise.NewProfileID(12345))
				Expect(err).ToNot(HaveOccurred())
				Expect(profile).ToNot(BeNil())
				Expect(profile.ID.Get()).To(Equal(int64(12345)))
				Expect(profile.Name).To(Equal("John Doe"))
			})
		})

		Context("with zero profile ID", func() {
			It("should return a rejection without calling the API", func() {
				_, err := client.GetProfile(context.Background(), wise.NewProfileID(0))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetExchangeRate", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/rates", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Query().Get("source")).To(Equal("EUR"))
					Expect(r.URL.Query().Get("target")).To(Equal("USD"))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.ExchangeRate{
						Source: "EUR", Target: "USD", Rate: 1.0857, Time: "2023-01-15T10:30:00Z",
					})
				})
			})

			It("should return the mapped rate", func() {
				rate, err := client.GetExchangeRate(
					context.Background(),
					wise.Currency("EUR"),
					wise.Currency("USD"),
					time.Time{},
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(rate).ToNot(BeNil())
				Expect(rate.Source).To(Equal(wise.Currency("EUR")))
				Expect(rate.Target).To(Equal(wise.Currency("USD")))
				Expect(rate.Rate).To(Equal(1.0857))
			})
		})

		Context("with missing source currency", func() {
			It("should return a rejection without calling the API", func() {
				_, err := client.GetExchangeRate(
					context.Background(),
					wise.Currency(""),
					wise.Currency("USD"),
					time.Time{},
				)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetTransfer", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/transfers/16521632", func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.Transfer{
						ID:             16521632,
						Status:         "delivered",
						TargetAccount:  98765432,
						SourceCurrency: "EUR",
						SourceValue:    1000,
						TargetCurrency: "USD",
						TargetValue:    1085.70,
						Rate:           1.0857,
						Created:        "2023-11-24 10:47:49",
					})
				})
			})

			It("should return the mapped transfer", func() {
				transfer, err := client.GetTransfer(context.Background(), wise.NewTransferID(16521632))
				Expect(err).ToNot(HaveOccurred())
				Expect(transfer).ToNot(BeNil())
				Expect(transfer.ID.Get()).To(Equal(int64(16521632)))
				Expect(transfer.Status).To(Equal(wise.TransferStatusDelivered))
				Expect(transfer.RecipientID.Get()).To(Equal(int64(98765432)))
			})
		})
	})

	Describe("CreateQuote", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v3/profiles/12345/quotes", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					var body map[string]any
					Expect(json.UnmarshalRead(r.Body, &body)).To(Succeed())
					Expect(body["sourceCurrency"]).To(Equal("EUR"))
					Expect(body["targetCurrency"]).To(Equal("USD"))
					Expect(body["sourceAmount"]).To(Equal(10.0))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.Quote{
						ID:             "11144c35-9fe8-4c32-b7fd-d05c2a7734bf",
						SourceCurrency: "EUR",
						TargetCurrency: "USD",
						SourceAmount:   10,
						TargetAmount:   10.86,
						PayOut:         "BANK_TRANSFER",
						Rate:           1.0857,
						CreatedTime:    "2023-01-15T10:30:00Z",
						ExpirationTime: "2023-01-15T11:00:00Z",
						Status:         "ACTIVE",
					})
				})
			})

			It("should create and map the quote", func() {
				quote, err := client.CreateQuote(
					context.Background(),
					wise.NewProfileID(12345),
					wise.CreateQuoteRequest{
						SourceCurrency: wise.Currency("EUR"),
						TargetCurrency: wise.Currency("USD"),
						SourceAmount:   &wise.Money{Cents: 1000, Currency: wise.Currency("EUR")},
						PayOut:         wise.PayOutBankTransfer,
					},
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(quote).ToNot(BeNil())
				Expect(quote.ID.Get()).To(Equal("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"))
				Expect(quote.Source.Cents).To(Equal(int64(1000)))
				Expect(quote.Profile.Get()).To(Equal(int64(12345)))
			})
		})
	})

	Describe("GetQuote", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v3/profiles/12345/quotes/11144c35-9fe8-4c32-b7fd-d05c2a7734bf",
					func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, raw.Quote{
							ID:             "11144c35-9fe8-4c32-b7fd-d05c2a7734bf",
							SourceCurrency: "EUR",
							TargetCurrency: "USD",
							SourceAmount:   10,
							TargetAmount:   10.86,
							PayOut:         "BANK_TRANSFER",
							Rate:           1.0857,
							CreatedTime:    "2023-01-15T10:30:00Z",
							ExpirationTime: "2023-01-15T11:00:00Z",
							Status:         "ACTIVE",
						})
					})
			})

			It("should return the mapped quote", func() {
				quote, err := client.GetQuote(
					context.Background(),
					wise.NewProfileID(12345),
					wise.NewQuoteID("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"),
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(quote).ToNot(BeNil())
				Expect(quote.ID.Get()).To(Equal("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"))
			})
		})
	})

	Describe("ListRecipients", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/accounts", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Query().Get("profile")).To(Equal("12345"))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, []raw.Recipient{
						{
							ID: 98765432, AccountHolderName: "Jane Doe",
							Currency: "GBP", Country: "GB", Type: "sort_code",
							Details: map[string]any{
								"sortCode":      "040075",
								"accountNumber": "37778842",
							},
							Active: true,
						},
					})
				})
			})

			It("should return mapped recipients", func() {
				recipients, err := client.ListRecipients(context.Background(), wise.ListRecipientsRequest{
					ProfileID: wise.NewProfileID(12345),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(recipients).To(HaveLen(1))
				Expect(recipients[0].ID.Get()).To(Equal(int64(98765432)))
				Expect(recipients[0].AccountHolderName).To(Equal("Jane Doe"))
				Expect(recipients[0].Details["sortCode"]).To(Equal("040075"))
			})
		})
	})

	Describe("GetRecipient", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/accounts/98765432", func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.Recipient{
						ID: 98765432, AccountHolderName: "Jane Doe",
						Currency: "GBP", Country: "GB", Type: "sort_code",
						Details: map[string]any{
							"sortCode":      "040075",
							"accountNumber": "37778842",
						},
						Active: true,
					})
				})
			})

			It("should return the mapped recipient", func() {
				recipient, err := client.GetRecipient(context.Background(), wise.NewRecipientID(98765432))
				Expect(err).ToNot(HaveOccurred())
				Expect(recipient).ToNot(BeNil())
				Expect(recipient.ID.Get()).To(Equal(int64(98765432)))
			})
		})
	})

	Describe("CreateRecipient", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/accounts", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					var body map[string]any
					Expect(json.UnmarshalRead(r.Body, &body)).To(Succeed())
					Expect(body["currency"]).To(Equal("GBP"))
					Expect(body["type"]).To(Equal("sort_code"))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.Recipient{
						ID: 98765432, AccountHolderName: "Jane Doe",
						Currency: "GBP", Country: "GB", Type: "sort_code",
						Details: map[string]any{
							"sortCode":      "040075",
							"accountNumber": "37778842",
						},
						Active: true,
					})
				})
			})

			It("should create and map the recipient", func() {
				recipient, err := client.CreateRecipient(context.Background(), wise.CreateRecipientRequest{
					ProfileID:         wise.NewProfileID(12345),
					Currency:          wise.Currency("GBP"),
					Type:              "sort_code",
					AccountHolderName: "Jane Doe",
					Details: map[string]string{
						"sortCode":      "040075",
						"accountNumber": "37778842",
					},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(recipient).ToNot(BeNil())
				Expect(recipient.ID.Get()).To(Equal(int64(98765432)))
			})
		})
	})

	Describe("CreateTransfer", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/transfers", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					var body map[string]any
					Expect(json.UnmarshalRead(r.Body, &body)).To(Succeed())
					Expect(body["targetAccount"]).To(Equal(float64(98765432)))
					Expect(body["quoteUuid"]).To(Equal("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"))
					Expect(body["customerTransactionId"]).To(Equal("22244c35-9fe8-4c32-b7fd-d05c2a7734bf"))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.Transfer{
						ID:             16521633,
						Status:         "incoming_payment_waiting",
						TargetAccount:  98765432,
						SourceCurrency: "EUR",
						SourceValue:    1000,
						TargetCurrency: "USD",
						TargetValue:    1085.70,
						Rate:           1.0857,
						Created:        "2023-11-24 10:47:49",
					})
				})
			})

			It("should create and map the transfer", func() {
				transfer, err := client.CreateTransfer(context.Background(), wise.CreateTransferRequest{
					QuoteID:               wise.NewQuoteID("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"),
					TargetAccount:         wise.NewRecipientID(98765432),
					CustomerTransactionID: "22244c35-9fe8-4c32-b7fd-d05c2a7734bf",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(transfer).ToNot(BeNil())
				Expect(transfer.ID.Get()).To(Equal(int64(16521633)))
				Expect(transfer.Status).To(Equal(wise.TransferStatusIncomingPaymentWaiting))
			})
		})
	})

	Describe("CancelTransfer", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/transfers/16521634/cancel", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal(http.MethodPut))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.Transfer{
						ID:             16521634,
						Status:         "cancelled",
						TargetAccount:  98765432,
						SourceCurrency: "EUR",
						SourceValue:    1000,
						TargetCurrency: "USD",
						TargetValue:    1085.70,
						Rate:           1.0857,
						Created:        "2023-11-24 10:47:49",
					})
				})
			})

			It("should cancel and map the transfer", func() {
				transfer, err := client.CancelTransfer(context.Background(), wise.NewTransferID(16521634))
				Expect(err).ToNot(HaveOccurred())
				Expect(transfer).ToNot(BeNil())
				Expect(transfer.ID.Get()).To(Equal(int64(16521634)))
				Expect(transfer.Status).To(Equal(wise.TransferStatusCancelled))
			})
		})

		Context("with zero transfer ID", func() {
			It("should return a rejection without calling the API", func() {
				_, err := client.CancelTransfer(context.Background(), wise.TransferID{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("transferID is required"))
			})
		})
	})

	Describe("FundTransfer", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/profiles/12345/transfers/16521634/payments",
					func(w http.ResponseWriter, r *http.Request) {
						Expect(r.Method).To(Equal(http.MethodPost))

						body, readErr := io.ReadAll(r.Body)
						Expect(readErr).To(Succeed())
						Expect(body).To(BeEmpty())

						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, raw.FundingResponse{
							Type:                 "BALANCE",
							Status:               "COMPLETED",
							BalanceTransactionID: new(int64(987654321)),
						})
					})
			})

			It("should fund and map the result", func() {
				result, err := client.FundTransfer(
					context.Background(),
					wise.NewProfileID(12345),
					wise.NewTransferID(16521634),
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Type).To(Equal(wise.FundingTypeBalance))
				Expect(result.Status).To(Equal(wise.FundingStatusCompleted))
				Expect(result.BalanceTransactionID).ToNot(BeNil())
				Expect(result.BalanceTransactionID.Get()).To(Equal(int64(987654321)))
			})
		})

		Context("when the API rejects the funding", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/profiles/12345/transfers/16521635/payments",
					func(w http.ResponseWriter, r *http.Request) {
						Expect(r.Method).To(Equal(http.MethodPost))

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = json.MarshalWrite(w, raw.FundingResponse{
							Type:         "BALANCE",
							Status:       "REJECTED",
							ErrorCode:    "balance.insufficient-funds",
							ErrorMessage: "Not enough funds",
						})
					})
			})

			It("should map the rejection into the result", func() {
				result, err := client.FundTransfer(
					context.Background(),
					wise.NewProfileID(12345),
					wise.NewTransferID(16521635),
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal(wise.FundingStatusRejected))
				Expect(result.ErrorCode).To(Equal(wise.FundingErrorCodeBalanceInsufficientFunds))
				Expect(result.ErrorMessage).To(Equal("Not enough funds"))
			})
		})

		Context("with zero profile ID", func() {
			It("should return a rejection without calling the API", func() {
				_, err := client.FundTransfer(
					context.Background(),
					wise.ProfileID{},
					wise.NewTransferID(16521634),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("profileID is required"))
			})
		})

		Context("with zero transfer ID", func() {
			It("should return a rejection without calling the API", func() {
				_, err := client.FundTransfer(
					context.Background(),
					wise.NewProfileID(12345),
					wise.TransferID{},
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("transferID is required"))
			})
		})
	})

	Describe("GetDeliveryEstimate", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/delivery-estimates/16521634", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Query().Get("timezone")).To(Equal("Asia/Singapore"))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, raw.DeliveryEstimate{
						EstimatedDeliveryDate:          "2018-01-10T12:15:00.000+0000",
						FormattedEstimatedDeliveryDate: "in seconds",
					})
				})
			})

			It("should return the mapped estimate", func() {
				estimate, err := client.GetDeliveryEstimate(
					context.Background(),
					wise.NewTransferID(16521634),
					"Asia/Singapore",
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(estimate).ToNot(BeNil())
				Expect(estimate.EstimatedDeliveryDate.UTC()).
					To(Equal(time.Date(2018, time.January, 10, 12, 15, 0, 0, time.UTC)))
				Expect(estimate.FormattedEstimatedDeliveryDate).To(Equal("in seconds"))
			})
		})

		Context("with zero transfer ID", func() {
			It("should return a rejection without calling the API", func() {
				_, err := client.GetDeliveryEstimate(context.Background(), wise.TransferID{}, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("transferID is required"))
			})
		})
	})

	Describe("ValidateTransferRequirements", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v1/transfer-requirements", func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					var reqBody map[string]any
					Expect(json.UnmarshalRead(r.Body, &reqBody)).To(Succeed())
					Expect(reqBody["targetAccount"]).To(Equal(float64(98765432)))
					Expect(reqBody["quoteUuid"]).To(Equal("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"))
					Expect(reqBody["originatorLegalEntityType"]).To(Equal("PRIVATE"))

					details, ok := reqBody["details"].(map[string]any)
					Expect(ok).To(BeTrue())
					Expect(details["reference"]).To(Equal("Invoice 2026-001"))

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, []raw.TransferRequirement{
						{
							Type: "transfer",
							Fields: []raw.TransferRequirementForm{
								{
									Name: "Transfer reference",
									Group: []raw.TransferRequirementField{
										{
											Key: "reference", Name: "Transfer reference", Type: "text",
											RefreshRequirementsOnChange: false, Required: false,
											MaxLength:        new(int32(10)),
											ValidationRegexp: new("[a-zA-Z0-9- ]*"),
										},
									},
								},
								{
									Name: "Transfer purpose",
									Group: []raw.TransferRequirementField{
										{
											Key: "transferPurpose", Name: "Transfer purpose", Type: "select",
											RefreshRequirementsOnChange: true, Required: true,
											ValuesAllowed: []raw.TransferRequirementValue{
												{
													Key:  "verification.transfers.purpose.pay.bills",
													Name: "Rent or other property expenses",
												},
											},
										},
									},
								},
							},
						},
					})
				})
			})

			It("should validate and map the dynamic form", func() {
				requirements, err := client.ValidateTransferRequirements(
					context.Background(),
					wise.ValidateTransferRequirementsRequest{
						TargetAccount:             wise.NewRecipientID(98765432),
						QuoteID:                   wise.NewQuoteID("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"),
						OriginatorLegalEntityType: "PRIVATE",
						Details: wise.TransferRequirementsDetails{
							Reference: "Invoice 2026-001",
						},
					},
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(requirements).To(HaveLen(1))
				Expect(requirements[0].Type).To(Equal("transfer"))
				Expect(requirements[0].Fields).To(HaveLen(2))

				reference := requirements[0].Fields[0].Group[0]
				Expect(reference.Key).To(Equal("reference"))
				Expect(reference.MaxLength).ToNot(BeNil())
				Expect(*reference.MaxLength).To(Equal(int32(10)))
				Expect(*reference.ValidationRegexp).To(Equal("[a-zA-Z0-9- ]*"))

				purpose := requirements[0].Fields[1].Group[0]
				Expect(purpose.Required).To(BeTrue())
				Expect(purpose.RefreshRequirementsOnChange).To(BeTrue())
				Expect(purpose.ValuesAllowed).To(HaveLen(1))
				Expect(purpose.ValuesAllowed[0].Key).To(Equal("verification.transfers.purpose.pay.bills"))
			})
		})

		Context("with missing quote ID", func() {
			It("should return a rejection without calling the API", func() {
				_, err := client.ValidateTransferRequirements(
					context.Background(),
					wise.ValidateTransferRequirementsRequest{
						TargetAccount: wise.NewRecipientID(98765432),
					},
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("quoteUuid is required"))
			})
		})
	})

	Describe("Quote payment options", func() {
		Context("when a quote includes paymentOptions and notices", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v3/profiles/12345/quotes/33344c35-9fe8-4c32-b7fd-d05c2a7734bf",
					func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						_ = json.MarshalWrite(w, raw.Quote{
							ID:             "33344c35-9fe8-4c32-b7fd-d05c2a7734bf",
							SourceCurrency: "GBP", TargetCurrency: "USD",
							SourceAmount: 100, TargetAmount: 129.24,
							PayOut: "BANK_TRANSFER", PayIn: "BALANCE",
							Rate:        1.30445,
							CreatedTime: "2023-01-15T10:30:00Z", ExpirationTime: "2023-01-15T11:00:00Z",
							Status:   "PENDING",
							Profile:  12345,
							RateType: "FIXED", ProvidedAmountType: "SOURCE",
							GuaranteedTargetAmountAllowed: true, GuaranteedTargetAmount: false,
							PaymentOptions: []raw.QuotePaymentOption{
								{
									Disabled:                   false,
									EstimatedDelivery:          "2023-01-16T12:30:00Z",
									FormattedEstimatedDelivery: "by Jan 16",
									Fee: raw.QuoteFee{
										TransferWise: 3.04, PayIn: 0, Discount: 2.27, Partner: 0, Total: 0.77,
									},
									SourceAmount: 100, TargetAmount: 129.24,
									SourceCurrency: "GBP", TargetCurrency: "USD",
									PayIn: "BALANCE", PayOut: "BANK_TRANSFER",
									PayInProduct: "CHEAP", FeePercentage: 0.0092,
								},
							},
							Notices: []raw.QuoteNotice{
								{
									Text: "You can have a maximum of 3 open transfers with a guaranteed rate.",
									Type: "WARNING",
								},
							},
						})
					})
			})

			It("should map payment options and notices", func() {
				quote, err := client.GetQuote(
					context.Background(),
					wise.NewProfileID(12345),
					wise.NewQuoteID("33344c35-9fe8-4c32-b7fd-d05c2a7734bf"),
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(quote).ToNot(BeNil())
				Expect(quote.RateType).To(Equal(wise.QuoteRateTypeFixed))
				Expect(quote.ProvidedAmountType).To(Equal(wise.QuoteProvidedAmountTypeSource))
				Expect(quote.GuaranteedTargetAmountAllowed).To(BeTrue())
				Expect(quote.GuaranteedTargetAmount).To(BeFalse())

				Expect(quote.PaymentOptions).To(HaveLen(1))
				option := quote.PaymentOptions[0]
				Expect(option.Disabled).To(BeFalse())
				Expect(option.FormattedEstimatedDelivery).To(Equal("by Jan 16"))
				Expect(option.Fee.Total).To(Equal(0.77))
				Expect(option.Source.Cents).To(Equal(int64(10000)))
				Expect(option.PayIn).To(Equal(wise.PayInBalance))
				Expect(option.PayOut).To(Equal(wise.PayOutBankTransfer))
				Expect(option.PayInProduct).To(Equal("CHEAP"))

				Expect(quote.Notices).To(HaveLen(1))
				Expect(quote.Notices[0].Type).To(Equal(wise.QuoteNoticeTypeWarning))
				Expect(quote.Notices[0].Link).To(BeNil())
			})
		})
	})

	Describe("Retry", func() {
		Context("on 429 with Retry-After header", func() {
			var callCount int

			BeforeEach(func() {
				callCount = 0

				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
					callCount++
					if callCount <= 3 {
						w.Header().Set("Retry-After", "0")
						w.WriteHeader(http.StatusTooManyRequests)
						_, _ = w.Write([]byte(
							`{"errors":[{"code":"RATE_LIMITED","message":"Too many requests"}]}`,
						))

						return
					}

					profiles := []raw.Profile{
						{
							ID: 1, Type: "PERSONAL", FirstName: "Test", LastName: "User",
							Email: "test@test.com", CreatedAt: "2023-01-01T00:00:00Z",
						},
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.MarshalWrite(w, profiles)
				})
			})

			It("should retry on 429 and eventually succeed", func() {
				_, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
