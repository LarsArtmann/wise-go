package wise_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larsartmann/wise-go"
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
) wise.Balance {
	return wise.Balance{
		ID: id, Currency: currency, Type: "STANDARD", Name: name,
		InvestmentState: investmentState, Visible: visible,
		Amount:         wise.BalanceAmount{Value: amount, Currency: currency},
		ReservedAmount: wise.BalanceAmount{Value: reserved, Currency: currency},
		CreationTime:   created,
	}
}

func testTx(id, txType string, amount float64) wise.StatementTransaction {
	return wise.StatementTransaction{
		TransactionID: id, Date: "2023-01-15 14:30:00",
		Amount:    wise.BalanceAmount{Value: amount, Currency: "EUR"},
		TotalFees: wise.BalanceAmount{Value: 0, Currency: "EUR"},
		Details:   wise.TransactionDetails{Type: txType},
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

func testProfiles(id int64, firstName, lastName, email string) []wise.Profile {
	return []wise.Profile{
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

func personalProfile(id int64, firstName, lastName, email, createdAt string) wise.Profile {
	return wise.Profile{
		ID: id, Type: "PERSONAL",
		FirstName: firstName, LastName: lastName,
		Email: email, CreatedAt: createdAt,
	}
}

func expectProfileID(profiles []wise.ProfileResult, idx int, expectedID int64) {
	Expect(profiles[idx].ID.Get()).To(Equal(expectedID))
}

func expectBalanceAmountCents(
	balances []wise.BalanceResult, idx int, expectedAmount, expectedReserved int64,
) {
	Expect(balances[idx].AmountCents).To(Equal(expectedAmount))
	Expect(balances[idx].ReservedCents).To(Equal(expectedReserved))
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

func listBalances(ctx context.Context, client *wise.Client) ([]wise.BalanceResult, error) {
	//nolint:wrapcheck // transparent pass-through; tests assert on the original error.
	return client.ListBalances(ctx, wise.NewProfileID(12345))
}

func getBalance(
	ctx context.Context, client *wise.Client, balanceID int64,
) (*wise.BalanceResult, error) {
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
			Currency:  "EUR",
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
					_ = json.NewEncoder(w).Encode(
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

				var authErr *wise.AuthError
				Expect(errors.As(err, &authErr)).To(BeTrue())
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
					profiles := []wise.Profile{
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
					_ = json.NewEncoder(w).Encode(profiles)
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

		Context("with unknown profile type", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
					profiles := []wise.Profile{
						{ID: 1, Type: "UNKNOWN_TYPE", CreatedAt: "2023-01-15T10:30:00Z"},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(profiles)
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
					func(w http.ResponseWriter, _ *http.Request) {
						balances := []wise.Balance{
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
						_ = json.NewEncoder(w).Encode(balances)
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
				Expect(balances[0].Currency).To(Equal("EUR"))
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
						balances := []wise.Balance{
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
						_ = json.NewEncoder(w).Encode(balances)
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
					balances := []wise.Balance{
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
					_ = json.NewEncoder(w).Encode(balances)
				},
			)
		})

		It("should return the specific balance", func() {
			balance, err := getBalance(context.Background(), client, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(balance.ID.Get()).To(Equal(int64(100)))
			Expect(balance.AmountCents).To(Equal(int64(100000)))
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

						response := wise.StatementResponse{
							Transactions: []wise.StatementTransaction{
								{
									TransactionID: "tx-001",
									Date:          "2023-01-15 14:30:00",
									Amount: wise.BalanceAmount{
										Value:    -50.00,
										Currency: "EUR",
									},
									TotalFees: wise.BalanceAmount{
										Value:    0.50,
										Currency: "EUR",
									},
									ReferenceNumber: "REF-001",
									Details: wise.TransactionDetails{
										Type:         "CARD_PAYMENT",
										Description:  "Coffee Shop",
										Category:     "food",
										MerchantName: "Starbucks",
									},
								},
								{
									TransactionID: "tx-002",
									Date:          "2023-01-20 10:00:00",
									Amount: wise.BalanceAmount{
										Value:    1000.00,
										Currency: "EUR",
									},
									TotalFees:       wise.BalanceAmount{Value: 0, Currency: "EUR"},
									ReferenceNumber: "REF-002",
									Details: wise.TransactionDetails{
										Type:        "TRANSFER",
										Description: "Salary Payment",
										Category:    "income",
									},
								},
							},
							EndOfStatementBalance: wise.BalanceAmount{
								Value:    950.50,
								Currency: "EUR",
							},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(response)
					},
				)
			})

			It("should return transactions", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions).To(HaveLen(2))
				Expect(resp.HasMore).To(BeFalse())
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
				Expect(resp.Transactions[0].TotalCents).To(Equal(int64(-5000)))
				Expect(resp.Transactions[0].AmountCents).To(Equal(int64(5000)))
				Expect(resp.Transactions[0].FeesCents).To(Equal(int64(50)))
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
		})

		Context("with transaction type edge cases", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						response := wise.StatementResponse{
							Transactions: []wise.StatementTransaction{
								testTx("tx-refund", "CARD_REFUND", 25.00),
								testTx("tx-exchange", "CONVERSION", -100.00),
								testTx("tx-fee", "FEE", -0.50),
								testTx("tx-payment", "PAYMENT", -200.00),
							},
						}
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(response)
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

				var notFoundErr *wise.NotFoundError
				Expect(errors.As(err, &notFoundErr)).To(BeTrue())
			})
		})

		Context("with running balance and exchange details", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						response := wise.StatementResponse{
							Transactions: []wise.StatementTransaction{
								{
									TransactionID: "tx-exch",
									Date:          "2023-01-15 14:30:00",
									Amount:        wise.BalanceAmount{Value: -100.00, Currency: "EUR"},
									TotalFees:     wise.BalanceAmount{Value: 0, Currency: "EUR"},
									RunningBalance: wise.BalanceAmount{
										Value: 5000.00, Currency: "EUR",
									},
									ReferenceNumber: "REF-EXCH",
									Details:         wise.TransactionDetails{Type: "CONVERSION"},
									ExchangeDetails: &wise.ExchangeDetails{
										FromAmount:   wise.BalanceAmount{Value: 100.00, Currency: "EUR"},
										ToAmount:     wise.BalanceAmount{Value: 108.50, Currency: "USD"},
										Rate:         1.085,
										FromCurrency: "EUR",
										ToCurrency:   "USD",
									},
								},
							},
						}
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(response)
					},
				)
			})

			It("should map running balance", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].RunningBalanceCents).To(Equal(int64(500000)))
				Expect(resp.Transactions[0].RunningBalanceCurrency).To(Equal("EUR"))
			})

			It("should map exchange details", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())

				exch := resp.Transactions[0].Exchange
				Expect(exch).ToNot(BeNil())
				Expect(exch.FromCents).To(Equal(int64(10000)))
				Expect(exch.FromCurrency).To(Equal("EUR"))
				Expect(exch.ToCents).To(Equal(int64(10850)))
				Expect(exch.ToCurrency).To(Equal("USD"))
				Expect(exch.Rate).To(Equal(1.085))
			})
		})

		Context("with a cross-currency transaction", func() {
			BeforeEach(func() {
				mux.HandleFunc(
					"/v1/profiles/12345/balance-statements/100/statement.json",
					func(w http.ResponseWriter, _ *http.Request) {
						response := wise.StatementResponse{
							Transactions: []wise.StatementTransaction{
								{
									TransactionID: "tx-usd",
									Date:          "2023-01-15 14:30:00",
									Amount:        wise.BalanceAmount{Value: -50.00, Currency: "USD"},
									TotalFees:     wise.BalanceAmount{Value: 0, Currency: "USD"},
									RunningBalance: wise.BalanceAmount{
										Value: 200.00, Currency: "USD",
									},
									ReferenceNumber: "REF-USD",
									Details:         wise.TransactionDetails{Type: "CARD_PAYMENT"},
								},
							},
						}
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(response)
					},
				)
			})

			It("should use the transaction currency, not the request currency", func() {
				resp, err := client.ListTransactions(context.Background(), defaultListTxReq)
				Expect(err).ToNot(HaveOccurred())

				tx := resp.Transactions[0]
				Expect(tx.AmountCurrency).To(Equal("USD"))
				Expect(tx.TotalCurrency).To(Equal("USD"))
				Expect(tx.RunningBalanceCurrency).To(Equal("USD"))
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
					profiles := []wise.Profile{
						{
							ID: 1, Type: "PERSONAL", FirstName: "Test", LastName: "User",
							Email: "test@test.com", CreatedAt: "2023-01-01T00:00:00Z",
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(profiles)
				})
			})

			It("should retry on 429 and eventually succeed", func() {
				_, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
