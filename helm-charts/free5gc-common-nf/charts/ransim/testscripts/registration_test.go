package test_test

import (
	"encoding/hex"
	"gofree5gc/lib/CommonConsumerTestData/UDM/TestGenAuthData"
	"gofree5gc/lib/CommonConsumerTestData/UDR/TestRegistrationProcedure"
	"gofree5gc/lib/MongoDBLibrary"
	"gofree5gc/lib/nas/nasMessage"
	"gofree5gc/lib/nas/nasTestpacket"
	"gofree5gc/lib/nas/nasType"
	"gofree5gc/lib/ngap"
	"gofree5gc/lib/ngap/ngapType"
	"gofree5gc/lib/openapi/models"

	// "gofree5gc/src/ausf/ausf_context"
	"gofree5gc/src/test"

	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
)

const ranIpAddr string = "{{ .Values.data.addr }}"

var (
	ueIdx   int
	icmpCnt int
	icmpDst string
	sst     int
	sd      string
)

var ueData test.UeData

func init() {
	flag.IntVar(&ueIdx, "ue-idx", 1, "UE index")
	flag.IntVar(&icmpCnt, "icmp-cnt", 5, "ICMP packet count to be sent")
	flag.StringVar(&icmpDst, "icmp-dst", "8.8.8.8", "Destination address of ICMP packets")
	flag.IntVar(&sst, "sst", 1, "SST of S-NSSAI")
	flag.StringVar(&sd, "sd", "010203", "SD of S-NSSAI")
	flag.Parse()

	// create ue
	supi := "imsi-20893000074" + strconv.Itoa(ueIdx+86)
	char_1 := supi[len(supi)-1:]
	char_2 := supi[len(supi)-2 : len(supi)-1]
	decoded, err := hex.DecodeString(char_1 + char_2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	buf := []uint8{0x01, 0x02, 0xf8, 0x39, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x47}
	buf = append(buf, []uint8(decoded)...)

	ueData = test.UeData{
		Supi:        supi,
		RanUeNgapId: int64(ueIdx),
		AmfUeNgapId: int64(ueIdx),
		MobileIdentity5GS: nasType.MobileIdentity5GS{
			Len:    12,
			Buffer: buf,
		},
		Sst: int32(sst),
		Sd:  sd,
	}
}

func getAuthSubscription() (authSubs models.AuthenticationSubscription) {
	authSubs.PermanentKey = &models.PermanentKey{
		PermanentKeyValue: TestGenAuthData.MilenageTestSet19.K,
	}
	authSubs.Opc = &models.Opc{
		OpcValue: TestGenAuthData.MilenageTestSet19.OPC,
	}
	authSubs.Milenage = &models.Milenage{
		Op: &models.Op{
			OpValue: TestGenAuthData.MilenageTestSet19.OP,
		},
	}
	authSubs.AuthenticationManagementField = "8000"

	authSubs.SequenceNumber = TestGenAuthData.MilenageTestSet19.SQN
	authSubs.AuthenticationMethod = models.AuthMethod__5_G_AKA
	return
}

func getAccessAndMobilitySubscriptionData() (amData models.AccessAndMobilitySubscriptionData) {
	return TestRegistrationProcedure.TestAmDataTable[TestRegistrationProcedure.FREE5GC_CASE]
}

func getSmfSelectionSubscriptionData() (smfSelData models.SmfSelectionSubscriptionData) {
	return TestRegistrationProcedure.TestSmfSelDataTable[TestRegistrationProcedure.FREE5GC_CASE]
}

func getAmPolicyData() (amPolicyData models.AmPolicyData) {
	return TestRegistrationProcedure.TestAmPolicyDataTable[TestRegistrationProcedure.FREE5GC_CASE]
}

func checksum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data) - 1
	)

	for i := 0; i < length; i += 2 {
		sum += uint32(data[i]) << 8
		sum += uint32(data[i+1])
	}

	if len(data)%2 == 1 {
		sum += uint32(data[length]) << 8
	}

	sum += sum >> 16

	return ^uint16(sum)
}

func buildGTPHeader(teid uint32, seq uint16) ([]byte, error) {
	const length uint16 = 52
	gtpheader := &layers.GTPv1U{
		Version:             1,
		ProtocolType:        1,
		Reserved:            0,
		ExtensionHeaderFlag: false,
		SequenceNumberFlag:  true,
		NPDUFlag:            false,
		MessageType:         255,
		MessageLength:       length,
		TEID:                teid,
		SequenceNumber:      seq,
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err := gtpheader.SerializeTo(buf, opts)

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Registration
func TestRegistration(t *testing.T) {
	var n int
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)

	// set MongoDB
	MongoDBLibrary.SetMongoDB("free5gc", "mongodb://{{ .Values.global.dbServiceDomain }}:27017")

	// RAN connect to AMF
	//// HELM: Remove spaces in substitution to avoid unclosed action error
	//// It seems to only occur when multiple substitution in one line
	conn, err := connectToAmf("{{.Values.amf.ngap.addr}}", "{{.Values.data.addr}}", 38412, 9487)
	assert.Nil(t, err)

	// send NGSetupRequest Msg
	sendMsg, err = test.GetNGSetupRequest([]byte("\x00\x01\x02"), 24, "free5gc")
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	// receive NGSetupResponse Msg
	n, err = conn.Read(recvMsg)
	assert.Nil(t, err)
	_, err = ngap.Decoder(recvMsg[:n])
	assert.Nil(t, err)

	// New UE
	ue := test.NewRanUeContext(ueData.Supi, 1, test.ALG_CIPHERING_128_NEA2, test.ALG_INTEGRITY_128_NIA2)
	// ue := test.NewRanUeContext("imsi-2089300007487", 1, test.ALG_CIPHERING_128_NEA0, test.ALG_INTEGRITY_128_NIA0)
	ue.AmfUeNgapId = ueData.AmfUeNgapId
	ue.RanUeNgapId = ueData.RanUeNgapId
	ue.AuthenticationSubs = getAuthSubscription()
	// insert UE data to MongoDB

	servingPlmnId := "20893"
	test.InsertAuthSubscriptionToMongoDB(ue.Supi, ue.AuthenticationSubs)
	getData := test.GetAuthSubscriptionFromMongoDB(ue.Supi)
	assert.NotNil(t, getData)
	{
		amData := getAccessAndMobilitySubscriptionData()
		test.InsertAccessAndMobilitySubscriptionDataToMongoDB(ue.Supi, amData, servingPlmnId)
		getData := test.GetAccessAndMobilitySubscriptionDataFromMongoDB(ue.Supi, servingPlmnId)
		assert.NotNil(t, getData)
	}
	{
		smfSelData := getSmfSelectionSubscriptionData()
		test.InsertSmfSelectionSubscriptionDataToMongoDB(ue.Supi, smfSelData, servingPlmnId)
		getData := test.GetSmfSelectionSubscriptionDataFromMongoDB(ue.Supi, servingPlmnId)
		assert.NotNil(t, getData)
	}
	{
		amPolicyData := getAmPolicyData()
		test.InsertAmPolicyDataToMongoDB(ue.Supi, amPolicyData)
		getData := test.GetAmPolicyDataFromMongoDB(ue.Supi)
		assert.NotNil(t, getData)
	}

	// send InitialUeMessage(Registration Request)(imsi-2089300007487)
	mobileIdentity5GS := ueData.MobileIdentity5GS
	registrationRequest := nasTestpacket.GetRegistrationRequestWith5GMM(nasMessage.RegistrationType5GSInitialRegistration, mobileIdentity5GS, nil, nil)
	sendMsg, err = test.GetInitialUEMessage(ue.RanUeNgapId, registrationRequest, "")
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	// receive NAS Authentication Request Msg
	n, err = conn.Read(recvMsg)
	assert.Nil(t, err)
	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	assert.Nil(t, err)

	// Calculate for RES*
	nasPdu := test.GetNasPdu(ngapMsg.InitiatingMessage.Value.DownlinkNASTransport)
	assert.NotNil(t, nasPdu)
	rand := nasPdu.AuthenticationRequest.GetRANDValue()
	resStat := ue.DeriveRESstarAndSetKey(ue.AuthenticationSubs, rand[:], "5G:mnc093.mcc208.3gppnetwork.org")

	// send NAS Authentication Response
	pdu := nasTestpacket.GetAuthenticationResponse(resStat, "")
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	// receive NAS Security Mode Command Msg
	n, err = conn.Read(recvMsg)
	assert.Nil(t, err)
	_, err = ngap.Decoder(recvMsg[:n])
	assert.Nil(t, err)

	// send NAS Security Mode Complete Msg
	pdu = nasTestpacket.GetSecurityModeComplete(registrationRequest)
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu)
	assert.Nil(t, err)
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	// receive ngap Initial Context Setup Request Msg
	n, err = conn.Read(recvMsg)
	assert.Nil(t, err)
	_, err = ngap.Decoder(recvMsg[:n])
	assert.Nil(t, err)

	// send ngap Initial Context Setup Response Msg
	sendMsg, err = test.GetInitialContextSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId)
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	// send NAS Registration Complete Msg
	pdu = nasTestpacket.GetRegistrationComplete(nil)
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu)
	assert.Nil(t, err)
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)
	// send GetPduSessionEstablishmentRequest Msg

	sNssai := models.Snssai{
		Sst: ueData.Sst,
		Sd:  ueData.Sd,
	}
	pdu = nasTestpacket.GetUlNasTransport_PduSessionEstablishmentRequest(10, nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", &sNssai)
	pdu, err = test.EncodeNasPduWithSecurity(ue, pdu)
	assert.Nil(t, err)
	sendMsg, err = test.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	// recieve 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	n, err = conn.Read(recvMsg)
	assert.Nil(t, err)
	ngapPduMsg, err := ngap.Decoder(recvMsg[:n])
	assert.Nil(t, err)

	// parse UPF address from NGAP-PDU Session Resource Setup Request
	var upfAddr string
	// fmt.Printf("%+v\n", ngapPduMsg.InitiatingMessage.Value)
	iEList := ngapPduMsg.InitiatingMessage.Value.PDUSessionResourceSetupRequest.ProtocolIEs.List
	for i := 0; i < len(iEList); i++ {
		// fmt.Printf("%+v\n", iEList[i].Id)
		// fmt.Printf("%+v\n", iEList[i].Value)
		if iEList[i].Value.Present == ngapType.PDUSessionResourceSetupRequestIEsPresentPDUSessionResourceSetupListSUReq {
			sUReqList := iEList[i].Value.PDUSessionResourceSetupListSUReq.List
			for j := 0; j < len(sUReqList); j++ {
				// fmt.Printf("%+v\n", sUReqList[j])
				if sUReqList[j].PDUSessionID.Value == int64(10) {
					// fmt.Printf("%+v\n", sUReqList[j].PDUSessionResourceSetupRequestTransfer)
					// TODO: unmarshal aper.OctetString and get TransportLayerAddress in field ULNGUUPTNLInformation
					upfAddr = net.IP(sUReqList[j].PDUSessionResourceSetupRequestTransfer[9:13]).String()
					break
				}
			}
		}
	}

	// send 14. NGAP-PDU Session Resource Setup Response
	sendMsg, err = test.GetPDUSessionResourceSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId, ranIpAddr)
	assert.Nil(t, err)
	_, err = conn.Write(sendMsg)
	assert.Nil(t, err)

	// wait 1s
	time.Sleep(1 * time.Second)

	// RAN connect to UPF
	upfConn, err := connectToUpf(ranIpAddr, upfAddr, 2152, 2152)
	assert.Nil(t, err)
	fmt.Printf("UPF %s connected\n", upfAddr)

	// receive ICMP packets
	sentTime := make(map[uint16]time.Time)
	go func(sentTime *map[uint16]time.Time) {
		for {
			// receive icmp response
			n, err = upfConn.Read(recvMsg)
			assert.Nil(t, err)
			recvTime := time.Now()

			// decode packet
			var gtpu layers.GTPv1U
			var ip4 layers.IPv4
			var icmp4 layers.ICMPv4
			var payload gopacket.Payload
			parser := gopacket.NewDecodingLayerParser(layers.LayerTypeGTPv1U, &gtpu, &ip4, &icmp4, &payload)
			decoded := []gopacket.LayerType{}
			err = parser.DecodeLayers(recvMsg, &decoded)
			if err != nil {
				fmt.Println("  Error encountered:", err)
			}
			var ttl uint8
			for _, layerType := range decoded {
				switch layerType {
				case layers.LayerTypeIPv4:
					ttl = ip4.TTL
				case layers.LayerTypeICMPv4:
					elapsed := recvTime.Sub((*sentTime)[icmp4.Seq])
					fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.2f ms\n", len(payload), icmpDst, icmp4.Seq, ttl, float64(elapsed)/float64(time.Millisecond))
				}
			}
		}
	}(&sentTime)
	// send ICMP packets
	fmt.Printf("Send %d ICMP packet(s) to %s\n", icmpCnt, icmpDst)
	for i := 0; i < icmpCnt; i++ {
		// build GTP header
		gtpHdrBuf, err := buildGTPHeader(1, uint16(i))
		assert.Nil(t, err)

		// build IPv4 header
		ipv4hdr := ipv4.Header{
			Version:  4,
			Len:      20,
			Protocol: 1,
			Flags:    0,
			TotalLen: 48,
			TTL:      64,
			Src:      net.ParseIP("60.60.0.1").To4(),
			Dst:      net.ParseIP(icmpDst).To4(),
			ID:       1,
			Checksum: 0,
		}
		v4HdrBuf, err := ipv4hdr.Marshal()
		assert.Nil(t, err)

		// compute IP checksum
		ipv4hdr.Checksum = int(checksum(v4HdrBuf))
		v4HdrBuf, err = ipv4hdr.Marshal()
		assert.Nil(t, err)

		// build ICMP packet
		icmpData, err := hex.DecodeString("8c870d0000000000101112131415161718191a1b")
		assert.Nil(t, err)

		icmpPacket := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: 0, Seq: i,
				Data: icmpData,
			},
		}
		icmpBuf, err := icmpPacket.Marshal(nil)
		assert.Nil(t, err)

		// concat headers and payload
		gtpPacket := append(gtpHdrBuf, v4HdrBuf...)
		gtpPacket = append(gtpPacket, icmpBuf...)

		_, err = upfConn.Write(gtpPacket)
		assert.Nil(t, err)
		sentTime[uint16(i)] = time.Now()

		// wait one second interval
		time.Sleep(1 * time.Second)
	}

	// delete test data
	test.DelAuthSubscriptionToMongoDB(ue.Supi)
	test.DelAccessAndMobilitySubscriptionDataFromMongoDB(ue.Supi, servingPlmnId)
	test.DelSmfSelectionSubscriptionDataFromMongoDB(ue.Supi, servingPlmnId)

	// close Connection
	conn.Close()
}
