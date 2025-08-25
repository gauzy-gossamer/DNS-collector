package transformers

import (
	"strings"
	"testing"

	"github.com/dmachard/go-dnscollector/dnsutils"
	"github.com/dmachard/go-dnscollector/pkgconfig"
	"github.com/dmachard/go-logger"
)

func TestNormalize_LowercaseQname(t *testing.T) {
	// enable feature
	config := pkgconfig.GetFakeConfigTransformers()
	config.Normalize.Enable = true
	config.Normalize.QnameLowerCase = true

	outChans := []chan dnsutils.DNSMessage{}

	// init the processor
	normTransformer := NewNormalizeTransform(config, logger.New(false), "test", 0, outChans)

	qname := "www.Google.Com"
	dm := dnsutils.GetFakeDNSMessage()
	dm.DNS.Qname = qname

	returnCode, err := normTransformer.QnameLowercase(&dm)
	if err != nil {
		t.Errorf("process transform err %s", err.Error())
	}

	if dm.DNS.Qname != NormAddress {
		t.Errorf("Qname to lowercase failed, got %s", dm.DNS.Qname)
	}
	if returnCode != ReturnKeep {
		t.Errorf("Return code is %v and not RETURN_KEEP (%v)", returnCode, ReturnKeep)
	}
}

func TestNormalize_RRLowercaseQname(t *testing.T) {
	// enable feature
	config := pkgconfig.GetFakeConfigTransformers()
	config.Normalize.Enable = true

	outChans := []chan dnsutils.DNSMessage{}

	// init the processor
	normTransformer := NewNormalizeTransform(config, logger.New(false), "test", 0, outChans)

	// create DNSMessage with answers
	rrqname := "www.RRGoogle.Com"
	dm := dnsutils.GetFakeDNSMessage()
	dm.DNS.Qname = "www.test.com"
	dm.DNS.DNSRRs.Answers = append(dm.DNS.DNSRRs.Answers, dnsutils.DNSAnswer{Name: rrqname})
	dm.DNS.DNSRRs.Nameservers = append(dm.DNS.DNSRRs.Nameservers, dnsutils.DNSAnswer{Name: rrqname})
	dm.DNS.DNSRRs.Records = append(dm.DNS.DNSRRs.Records, dnsutils.DNSAnswer{Name: rrqname})

	// process DNSMessage
	returnCode, err := normTransformer.RRLowercase(&dm)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if returnCode != ReturnKeep {
		t.Errorf("Return code is %v, want %v", returnCode, ReturnKeep)
	}

	// checks
	if dm.DNS.DNSRRs.Answers[0].Name != strings.ToLower(rrqname) {
		t.Errorf("RR Answers to lowercase failed, got %s", dm.DNS.DNSRRs.Answers[0].Name)
	}

	if dm.DNS.DNSRRs.Nameservers[0].Name != strings.ToLower(rrqname) {
		t.Errorf("RR Nameservers to lowercase failed, got %s", dm.DNS.DNSRRs.Nameservers[0].Name)
	}

	if dm.DNS.DNSRRs.Records[0].Name != strings.ToLower(rrqname) {
		t.Errorf("RR Records to lowercase failed, got %s", dm.DNS.DNSRRs.Records[0].Name)
	}
}

func TestNormalize_QuietText(t *testing.T) {
	// enable feature
	config := pkgconfig.GetFakeConfigTransformers()
	config.Normalize.Enable = true
	config.Normalize.QuietText = true

	outChans := []chan dnsutils.DNSMessage{}

	// init the processor
	norm := NewNormalizeTransform(config, logger.New(false), "test", 0, outChans)

	dm := dnsutils.GetFakeDNSMessage()
	norm.QuietText(&dm)

	if dm.DNSTap.Operation != "CQ" {
		t.Errorf("CQ expected: %s", dm.DNSTap.Operation)
	}

	if dm.DNS.Type != "Q" {
		t.Errorf("Q expected: %s", dm.DNS.Type)
	}
}

func TestNormalize_AddTLD(t *testing.T) {
	// enable feature
	config := pkgconfig.GetFakeConfigTransformers()
	config.Normalize.Enable = true
	config.Normalize.AddTld = true

	outChans := []chan dnsutils.DNSMessage{}

	// init the processor
	psl := NewNormalizeTransform(config, logger.New(false), "test", 0, outChans)

	tt := []struct {
		name  string
		qname string
		want  string
	}{
		{
			name:  "get tld",
			qname: "www.amazon.fr",
			want:  "fr",
		},
		{
			name:  "get tld insensitive",
			qname: "www.Google.Com",
			want:  "com",
		},
		{
			name:  "get tld with dot trailing",
			qname: "www.amazon.fr.",
			want:  "fr",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			dm := dnsutils.GetFakeDNSMessage()
			dm.DNS.Qname = tc.qname

			// psl.InitDNSMessage(&dm)

			psl.GetEffectiveTld(&dm)
			if dm.PublicSuffix.QnamePublicSuffix != tc.want {
				t.Errorf("Bad TLD, got: %s, expected: com", dm.PublicSuffix.QnamePublicSuffix)

			}
		})
	}
}

func TestNormalize_AddTldPlusOne(t *testing.T) {
	// enable feature
	config := pkgconfig.GetFakeConfigTransformers()
	config.Normalize.Enable = true
	config.Normalize.AddTld = true

	outChans := []chan dnsutils.DNSMessage{}

	// init the processor
	psl := NewNormalizeTransform(config, logger.New(false), "test", 0, outChans)

	tt := []struct {
		name  string
		qname string
		want  string
	}{
		{
			name:  "get tld",
			qname: "www.amazon.fr",
			want:  "amazon.fr",
		},
		{
			name:  "get tld insensitive",
			qname: "books.amazon.co.uk",
			want:  "amazon.co.uk",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			dm := dnsutils.GetFakeDNSMessage()
			dm.DNS.Qname = tc.qname

			// psl.InitDNSMessage(&dm)

			psl.GetEffectiveTldPlusOne(&dm)
			if dm.PublicSuffix.QnameEffectiveTLDPlusOne != tc.want {
				t.Errorf("Bad TLD, got: %s, expected: %s", dm.PublicSuffix.QnameEffectiveTLDPlusOne, tc.want)

			}
		})
	}
}

func TestNormalize_SuffixUnmanaged(t *testing.T) {
	// enable feature
	config := pkgconfig.GetFakeConfigTransformers()
	outChans := []chan dnsutils.DNSMessage{}

	// init the processor
	psl := NewNormalizeTransform(config, logger.New(true), "test", 0, outChans)

	dm := dnsutils.GetFakeDNSMessage()
	// https://publicsuffix.org/list/effective_tld_names.dat
	// // ===BEGIN ICANN DOMAINS===
	// ....
	// // ===END ICANN DOMAINS===
	// ===BEGIN PRIVATE DOMAINS===
	// ..
	dm.DNS.Qname = "play.googleapis.com"
	// // ===END PRIVATE DOMAINS===

	// psl.InitDNSMessage(&dm)
	psl.GetEffectiveTld(&dm)
	if dm.PublicSuffix.ManagedByICANN {
		t.Errorf("Qname %s should be private domains", dm.DNS.Qname)
	}
}

func TestNormalize_SuffixICANNManaged(t *testing.T) {
	// enable feature
	config := pkgconfig.GetFakeConfigTransformers()
	outChans := []chan dnsutils.DNSMessage{}

	// init the processor
	psl := NewNormalizeTransform(config, logger.New(true), "test", 0, outChans)

	dm := dnsutils.GetFakeDNSMessage()
	// https://publicsuffix.org/list/effective_tld_names.dat
	// // ===BEGIN ICANN DOMAINS===
	dm.DNS.Qname = "fr.wikipedia.org"
	// // ===END ICANN DOMAINS===
	// ===BEGIN PRIVATE DOMAINS===
	// ..
	// // ===END PRIVATE DOMAINS===

	// psl.InitDNSMessage(&dm)
	psl.GetEffectiveTld(&dm)
	if !dm.PublicSuffix.ManagedByICANN {
		t.Errorf("Qname %s should be ICANN managed", dm.DNS.Qname)
	}
}

// bench tests

func BenchmarkNormalize_GetEffectiveTld(b *testing.B) {
	config := pkgconfig.GetFakeConfigTransformers()
	channels := []chan dnsutils.DNSMessage{}

	subprocessor := NewNormalizeTransform(config, logger.New(false), "test", 0, channels)
	dm := dnsutils.GetFakeDNSMessage()
	dm.DNS.Qname = "en.wikipedia.org"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// subprocessor.InitDNSMessage(&dm)
		subprocessor.GetEffectiveTld(&dm)
	}
}

func BenchmarkNormalize_GetEffectiveTldPlusOne(b *testing.B) {
	config := pkgconfig.GetFakeConfigTransformers()
	channels := []chan dnsutils.DNSMessage{}

	subprocessor := NewNormalizeTransform(config, logger.New(false), "test", 0, channels)
	dm := dnsutils.GetFakeDNSMessage()
	dm.DNS.Qname = "en.wikipedia.org"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// subprocessor.InitDNSMessage(&dm)
		subprocessor.GetEffectiveTld(&dm)
	}
}

func BenchmarkNormalize_QnameLowercase(b *testing.B) {
	config := pkgconfig.GetFakeConfigTransformers()
	channels := []chan dnsutils.DNSMessage{}

	subprocessor := NewNormalizeTransform(config, logger.New(false), "test", 0, channels)
	dm := dnsutils.GetFakeDNSMessage()
	dm.DNS.Qname = "EN.Wikipedia.Org"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subprocessor.QnameLowercase(&dm)
	}
}

func BenchmarkNormalize_RRLowercase(b *testing.B) {
	config := pkgconfig.GetFakeConfigTransformers()
	channels := []chan dnsutils.DNSMessage{}

	transform := NewNormalizeTransform(config, logger.New(false), "test", 0, channels)

	name := "En.Tikipedia.Org"
	dm := dnsutils.GetFakeDNSMessage()
	dm.DNS.Qname = name
	dm.DNS.DNSRRs.Answers = append(dm.DNS.DNSRRs.Answers, dnsutils.DNSAnswer{Name: name})
	dm.DNS.DNSRRs.Nameservers = append(dm.DNS.DNSRRs.Nameservers, dnsutils.DNSAnswer{Name: name})
	dm.DNS.DNSRRs.Records = append(dm.DNS.DNSRRs.Records, dnsutils.DNSAnswer{Name: name})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transform.RRLowercase(&dm)
	}
}

func BenchmarkNormalize_QuietText(b *testing.B) {
	config := pkgconfig.GetFakeConfigTransformers()
	channels := []chan dnsutils.DNSMessage{}

	subprocessor := NewNormalizeTransform(config, logger.New(false), "test", 0, channels)
	dm := dnsutils.GetFakeDNSMessage()
	dm.DNS.Qname = "EN.Wikipedia.Org"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subprocessor.QuietText(&dm)
	}
}
