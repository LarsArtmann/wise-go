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

var _ = Describe("Wise Client", func() {
	var (
		server *httptest.Server
		mux    *http.ServeMux
		client *wise.Client
	)

	BeforeEach(func() {
		mux = http.NewServeMux()
		server = httptest.NewServer(mux)
		client = wise.New("test-api-key", wise.WithBaseURL(server.URL))
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
					profiles := []wise.Profile{
						{
							ID:        12345,
							Type:      "PERSONAL",
							FirstName: "John",
							LastName:  "Doe",
							Email:     "john@example.com",
							CreatedAt: "2023-01-15T10:30:00Z",
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(profiles)
				})
			})

			It("should not return error", func() {
				err := client.Authenticate(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with invalid credentials", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"errors":[{"code":"UNAUTHORIZED","message":"Invalid API key"}]}`))
				})
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
						{
							ID: 12345, Type: "PERSONAL",
							FirstName: "John", LastName: "Doe",
							Email: "john@example.com", CreatedAt: "2023-01-15T10:30:00Z",
						},
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

				Expect(profiles[0].ID).To(Equal(int64(12345)))
				Expect(profiles[0].Name).To(Equal("John Doe"))
				Expect(profiles[0].Type).To(Equal(wise.ProfileTypePersonal))
				Expect(profiles[0].Email).To(Equal("john@example.com"))
			})

			It("should map business profile correctly", func() {
				profiles, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())

				Expect(profiles[1].ID).To(Equal(int64(67890)))
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
				mux.HandleFunc("/v2/profiles", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"errors":[{"code":"UNAUTHORIZED","message":"Invalid API key"}]}`))
				})
			})

			It("should return error with API error details", func() {
				_, err := client.ListProfiles(context.Background())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("UNAUTHORIZED"))
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
				_, err := client.ListProfiles(context.Background())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unknown profile type"))
			})
		})
	})

	Describe("ListBalances", func() {
		Context("with valid API response", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v4/profiles/12345/balances", func(w http.ResponseWriter, _ *http.Request) {
					balances := []wise.Balance{
						{
							ID: 100, Currency: "EUR", Type: "STANDARD", Name: "Main Account",
							InvestmentState: "NOT_INVESTED", Visible: true,
							Amount:         wise.BalanceAmount{Value: 1234.56, Currency: "EUR"},
							ReservedAmount: wise.BalanceAmount{Value: 0.0, Currency: "EUR"},
							CreationTime:   "2023-01-01T00:00:00Z",
						},
						{
							ID: 200, Currency: "USD", Type: "STANDARD", Name: "USD Account",
							InvestmentState: "NOT_INVESTED", Visible: true,
							Amount:         wise.BalanceAmount{Value: 500.00, Currency: "USD"},
							ReservedAmount: wise.BalanceAmount{Value: 50.00, Currency: "USD"},
							CreationTime:   "2023-01-02T00:00:00Z",
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(balances)
				})
			})

			It("should return visible balances", func() {
				balances, err := client.ListBalances(context.Background(), 12345)
				Expect(err).ToNot(HaveOccurred())
				Expect(balances).To(HaveLen(2))
			})

			It("should map balance fields correctly", func() {
				balances, err := client.ListBalances(context.Background(), 12345)
				Expect(err).ToNot(HaveOccurred())

				Expect(balances[0].ID).To(Equal(int64(100)))
				Expect(balances[0].Currency).To(Equal("EUR"))
				Expect(balances[0].Type).To(Equal(wise.BalanceTypeStandard))
				Expect(balances[0].Name).To(Equal("Main Account"))
			})

			It("should convert amounts to cents", func() {
				balances, err := client.ListBalances(context.Background(), 12345)
				Expect(err).ToNot(HaveOccurred())

				Expect(balances[0].AmountCents).To(Equal(int64(123456)))
				Expect(balances[0].ReservedCents).To(Equal(int64(0)))
				Expect(balances[1].AmountCents).To(Equal(int64(50000)))
				Expect(balances[1].ReservedCents).To(Equal(int64(5000)))
			})
		})

		Context("with invisible balances", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v4/profiles/12345/balances", func(w http.ResponseWriter, _ *http.Request) {
					balances := []wise.Balance{
						{
							ID: 100, Currency: "EUR", Type: "STANDARD", Name: "Visible",
							InvestmentState: "NOT_INVESTED", Visible: true,
							Amount:         wise.BalanceAmount{Value: 100.00, Currency: "EUR"},
							ReservedAmount: wise.BalanceAmount{Value: 0, Currency: "EUR"},
							CreationTime:   "2023-01-01T00:00:00Z",
						},
						{
							ID: 200, Currency: "EUR", Type: "STANDARD", Name: "Invisible",
							InvestmentState: "NOT_INVESTED", Visible: false,
							Amount:         wise.BalanceAmount{Value: 200.00, Currency: "EUR"},
							ReservedAmount: wise.BalanceAmount{Value: 0, Currency: "EUR"},
							CreationTime:   "2023-01-01T00:00:00Z",
						},
						{
							ID: 300, Currency: "EUR", Type: "STANDARD", Name: "Investment",
							InvestmentState: "INVESTED", Visible: true,
							Amount:         wise.BalanceAmount{Value: 300.00, Currency: "EUR"},
							ReservedAmount: wise.BalanceAmount{Value: 0, Currency: "EUR"},
							CreationTime:   "2023-01-01T00:00:00Z",
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(balances)
				})
			})

			It("should filter out invisible and investment balances", func() {
				balances, err := client.ListBalances(context.Background(), 12345)
				Expect(err).ToNot(HaveOccurred())
				Expect(balances).To(HaveLen(1))
				Expect(balances[0].Name).To(Equal("Visible"))
			})
		})

		Context("with API error", func() {
			BeforeEach(func() {
				mux.HandleFunc("/v4/profiles/12345/balances", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"errors":[{"code":"SERVER_ERROR","message":"Internal server error"}]}`))
				})
			})

			It("should return error", func() {
				_, err := client.ListBalances(context.Background(), 12345)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetBalance", func() {
		BeforeEach(func() {
			mux.HandleFunc("/v4/profiles/12345/balances", func(w http.ResponseWriter, _ *http.Request) {
				balances := []wise.Balance{
					{
						ID: 100, Currency: "EUR", Type: "STANDARD", Name: "EUR Account",
						InvestmentState: "NOT_INVESTED", Visible: true,
						Amount:         wise.BalanceAmount{Value: 1000.00, Currency: "EUR"},
						ReservedAmount: wise.BalanceAmount{Value: 0, Currency: "EUR"},
						CreationTime:   "2023-01-01T00:00:00Z",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(balances)
			})
		})

		It("should return the specific balance", func() {
			balance, err := client.GetBalance(context.Background(), 12345, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(balance.ID).To(Equal(int64(100)))
			Expect(balance.AmountCents).To(Equal(int64(100000)))
		})

		It("should return error for non-existent balance", func() {
			_, err := client.GetBalance(context.Background(), 12345, 999)
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
						Expect(r.URL.Query().Get("currency")).To(Equal("EUR"))
						Expect(r.URL.Query().Get("intervalStart")).To(Equal("2023-01-01T00:00:00Z"))
						Expect(r.URL.Query().Get("intervalEnd")).To(Equal("2023-01-31T23:59:59Z"))

						response := wise.StatementResponse{
							Transactions: []wise.StatementTransaction{
								{
									TransactionID:   "tx-001",
									Date:            "2023-01-15 14:30:00",
									Amount:          wise.BalanceAmount{Value: -50.00, Currency: "EUR"},
									TotalFees:       wise.BalanceAmount{Value: 0.50, Currency: "EUR"},
									ReferenceNumber: "REF-001",
									Details: wise.TransactionDetails{
										Type:         "CARD_PAYMENT",
										Description:  "Coffee Shop",
										Category:     "food",
										MerchantName: "Starbucks",
									},
								},
								{
									TransactionID:   "tx-002",
									Date:            "2023-01-20 10:00:00",
									Amount:          wise.BalanceAmount{Value: 1000.00, Currency: "EUR"},
									TotalFees:       wise.BalanceAmount{Value: 0, Currency: "EUR"},
									ReferenceNumber: "REF-002",
									Details: wise.TransactionDetails{
										Type:        "TRANSFER",
										Description: "Salary Payment",
										Category:    "income",
									},
								},
							},
							EndOfStatementBalance: wise.BalanceAmount{Value: 950.50, Currency: "EUR"},
						}

						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(response)
					},
				)
			})

			It("should return transactions", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345,
					BalanceID: 100,
					Currency:  "EUR",
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions).To(HaveLen(2))
				Expect(resp.HasMore).To(BeFalse())
			})

			It("should map transaction ID correctly", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345,
					BalanceID: 100,
					Currency:  "EUR",
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].ID).To(Equal("tx-001"))
			})

			It("should classify card payment as card type", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345,
					BalanceID: 100,
					Currency:  "EUR",
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].Type).To(Equal(wise.TransactionTypeCard))
			})

			It("should classify transfer as transfer type", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345,
					BalanceID: 100,
					Currency:  "EUR",
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[1].Type).To(Equal(wise.TransactionTypeTransfer))
			})

			It("should convert amounts to cents correctly", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345,
					BalanceID: 100,
					Currency:  "EUR",
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())

				// First transaction: -50.00 → totalCents=-5000, amountCents=5000 (abs)
				Expect(resp.Transactions[0].TotalCents).To(Equal(int64(-5000)))
				Expect(resp.Transactions[0].AmountCents).To(Equal(int64(5000)))
				Expect(resp.Transactions[0].FeesCents).To(Equal(int64(50)))
			})

			It("should parse date correctly", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345,
					BalanceID: 100,
					Currency:  "EUR",
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].Date.Year()).To(Equal(2023))
				Expect(resp.Transactions[0].Date.Month()).To(Equal(time.January))
				Expect(resp.Transactions[0].Date.Day()).To(Equal(15))
			})

			It("should map description and merchant", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345,
					BalanceID: 100,
					Currency:  "EUR",
					From:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:        time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
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
								{
									TransactionID: "tx-refund", Date: "2023-01-15 14:30:00",
									Amount:    wise.BalanceAmount{Value: 25.00, Currency: "EUR"},
									TotalFees: wise.BalanceAmount{Value: 0, Currency: "EUR"},
									Details:   wise.TransactionDetails{Type: "CARD_REFUND"},
								},
								{
									TransactionID: "tx-exchange", Date: "2023-01-15 14:30:00",
									Amount:    wise.BalanceAmount{Value: -100.00, Currency: "EUR"},
									TotalFees: wise.BalanceAmount{Value: 0, Currency: "EUR"},
									Details:   wise.TransactionDetails{Type: "CONVERSION"},
								},
								{
									TransactionID: "tx-fee", Date: "2023-01-15 14:30:00",
									Amount:    wise.BalanceAmount{Value: -0.50, Currency: "EUR"},
									TotalFees: wise.BalanceAmount{Value: 0, Currency: "EUR"},
									Details:   wise.TransactionDetails{Type: "FEE"},
								},
								{
									TransactionID: "tx-payment", Date: "2023-01-15 14:30:00",
									Amount:    wise.BalanceAmount{Value: -200.00, Currency: "EUR"},
									TotalFees: wise.BalanceAmount{Value: 0, Currency: "EUR"},
									Details:   wise.TransactionDetails{Type: "PAYMENT"},
								},
							},
						}
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(response)
					},
				)
			})

			It("should classify CARD_REFUND with positive amount as refund", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345, BalanceID: 100, Currency: "EUR",
					From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[0].Type).To(Equal(wise.TransactionTypeRefund))
			})

			It("should classify CONVERSION as exchange", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345, BalanceID: 100, Currency: "EUR",
					From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[1].Type).To(Equal(wise.TransactionTypeExchange))
			})

			It("should classify FEE as fee", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345, BalanceID: 100, Currency: "EUR",
					From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Transactions[2].Type).To(Equal(wise.TransactionTypeFee))
			})

			It("should classify PAYMENT as payment", func() {
				resp, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345, BalanceID: 100, Currency: "EUR",
					From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
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
						_, _ = w.Write([]byte(`{"errors":[{"code":"NOT_FOUND","message":"Balance not found"}]}`))
					},
				)
			})

			It("should return NotFoundError", func() {
				_, err := client.ListTransactions(context.Background(), wise.ListTransactionsRequest{
					ProfileID: 12345, BalanceID: 100, Currency: "EUR",
					From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC),
				})
				Expect(err).To(HaveOccurred())

				var notFoundErr *wise.NotFoundError
				Expect(errors.As(err, &notFoundErr)).To(BeTrue())
			})
		})

		Context("with rate limit error", func() {
			BeforeEach(func() {
				callCount := 0
				mux.HandleFunc(
					"/v2/profiles",
					func(w http.ResponseWriter, _ *http.Request) {
						callCount++
						if callCount <= 3 {
							w.WriteHeader(http.StatusTooManyRequests)
							_, _ = w.Write([]byte(`{"errors":[{"code":"RATE_LIMITED","message":"Too many requests"}]}`))

							return
						}
						profiles := []wise.Profile{
							{ID: 1, Type: "PERSONAL", FirstName: "Test", LastName: "User", Email: "test@test.com", CreatedAt: "2023-01-01T00:00:00Z"},
						}
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(profiles)
					},
				)
			})

			It("should retry on 429 and eventually succeed", func() {
				_, err := client.ListProfiles(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
