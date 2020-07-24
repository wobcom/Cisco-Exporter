package environment

var (
	nxos1 = `
Fan:
------------------------------------------------------
Fan             Model                Hw         Status
------------------------------------------------------
Fan-1           N3K-C3048-FAN        --         ok
PS-1            N2200-PAC-400W       --         ok
PS-2            N2200-PAC-400W       --         ok


Temprature
-------------------------------------------------------------------------
Module  Sensor             MajorThresh   MinorThres   CurTemp     Status
                            (Celsius)     (Celsius)   (Celsius)
-------------------------------------------------------------------------
1       Back        (D0)    55            45          42           ok
2       Front Middle(D1)    65            57          57           ok
3       Front Right (D2)    60            52          45           ok
4       Front Left  (D3)    60            51          54           ok


Power Supply
Voltage: 12 Volts
-----------------------------------------------------------
PS  Model                Input Power       Current   Status
                         Type  (Watts)     (Amps)
-----------------------------------------------------------
1   N2200-PAC-400W       AC     396.00      33.00    ok
2   N2200-PAC-400W       AC     396.00      33.00    ok


Mod Model                   Power     Current     Power     Current     Status
                            Requested Requested   Allocated Allocated
                            (Watts)   (Amps)      (Watts)   (Amps)
--- ----------------------  -------   ----------  --------- ----------  ----------
1   N3K-C3048TP-1GE-SUP     319.20    29.10       349.20    29.10       powered-up

Power Usage Summary:
--------------------
Power Supply redundancy mode:                 Redundant
Power Supply redundancy operational mode:     Redundant

Total Power Capacity                              792.00 W

Power reserved for Supervisor(s)                  349.20 W
Power current used by Modeules                      0.00 W

                                                -------------
Total Power Available                             442.80 W
                                                -------------
`

	expectedNxos1Metrics = map[string]float64{
		prefix + "fan_operational_status_info{fan=Fan-1,hw=--,model=N3K-C3048-FAN,target=test.test}":           1,
		prefix + "fan_operational_status_info{fan=PS-1,hw=--,model=N2200-PAC-400W,target=test.test}":           1,
		prefix + "fan_operational_status_info{fan=PS-2,hw=--,model=N2200-PAC-400W,target=test.test}":           1,
		prefix + "powersupply_allocated_current_amps{mod=1,model=N3K-C3048TP-1GE-SUP,target=test.test}":        29.1,
		prefix + "powersupply_allocated_power_watts{mod=1,model=N3K-C3048TP-1GE-SUP,target=test.test}":         349.2,
		prefix + "powersupply_capacity_total_watts{target=test.test}":                                          792,
		prefix + "powersupply_current_amps{input_type=AC,model=N2200-PAC-400W,ps=1,target=test.test}":          33,
		prefix + "powersupply_current_amps{input_type=AC,model=N2200-PAC-400W,ps=2,target=test.test}":          33,
		prefix + "powersupply_operational_info{input_type=AC,model=N2200-PAC-400W,ps=1,target=test.test}":      1,
		prefix + "powersupply_operational_info{input_type=AC,model=N2200-PAC-400W,ps=2,target=test.test}":      1,
		prefix + "powersupply_power_watts{input_type=AC,model=N2200-PAC-400W,ps=1,target=test.test}":           396,
		prefix + "powersupply_power_watts{input_type=AC,model=N2200-PAC-400W,ps=2,target=test.test}":           396,
		prefix + "powersupply_redundancy_configured_info{target=test.test}":                                    1,
		prefix + "powersupply_redundancy_opererational_info{target=test.test}":                                 1,
		prefix + "powersupply_requested_current_amps{mod=1,model=N3K-C3048TP-1GE-SUP,target=test.test}":        29.100000,
		prefix + "powersupply_requested_power_watts{mod=1,model=N3K-C3048TP-1GE-SUP,target=test.test}":         319.200000,
		prefix + "powersupply_status_info{mod=1,model=N3K-C3048TP-1GE-SUP,status=powered-up,target=test.test}": 1,
		prefix + "powersupply_total_power_available_watts{target=test.test}":                                   442.8,
		prefix + "powersupply_voltage_volts{target=test.test}":                                                 12,
		prefix + "temperature_current_celsius{module=1,sensor=Back        (D0),target=test.test}":              42,
		prefix + "temperature_current_celsius{module=2,sensor=Front Middle(D1),target=test.test}":              57,
		prefix + "temperature_current_celsius{module=3,sensor=Front Right (D2),target=test.test}":              45,
		prefix + "temperature_current_celsius{module=4,sensor=Front Left  (D3),target=test.test}":              54,
		prefix + "temperature_major_threshold_celsius{module=1,sensor=Back        (D0),target=test.test}":      55,
		prefix + "temperature_major_threshold_celsius{module=2,sensor=Front Middle(D1),target=test.test}":      65,
		prefix + "temperature_major_threshold_celsius{module=3,sensor=Front Right (D2),target=test.test}":      60,
		prefix + "temperature_major_threshold_celsius{module=4,sensor=Front Left  (D3),target=test.test}":      60,
		prefix + "temperature_minor_threshold_celsius{module=1,sensor=Back        (D0),target=test.test}":      45,
		prefix + "temperature_minor_threshold_celsius{module=2,sensor=Front Middle(D1),target=test.test}":      57,
		prefix + "temperature_minor_threshold_celsius{module=3,sensor=Front Right (D2),target=test.test}":      52,
		prefix + "temperature_minor_threshold_celsius{module=4,sensor=Front Left  (D3),target=test.test}":      51,
	}

	nxos2 = `
Fan:
---------------------------------------------------------------------------
Fan             Model                Hw     Direction       Status
---------------------------------------------------------------------------
Fan1(sys_fan1)  NXA-FAN-30CFM-B      --     front-to-back   Ok 
Fan2(sys_fan2)  NXA-FAN-30CFM-B      --     front-to-back   Ok 
Fan3(sys_fan3)  NXA-FAN-30CFM-B      --     front-to-back   Ok 
Fan4(sys_fan4)  NXA-FAN-30CFM-B      --     front-to-back   Ok 
Fan_in_PS1      --                   --     front-to-back   Ok
Fan_in_PS2      --                   --     front-to-back   Ok
Fan Zone Speed: Zone 1: 0x80
Fan Air Filter : NotSupported


Power Supply:
Voltage: 12 Volts
Power                             Actual      Actual        Total
Supply    Model                   Output      Input      Capacity    Status
                                (Watts )    (Watts )     (Watts )
-------  -------------------  ----------  ----------  ----------  --------------
1        NXA-PAC-650W-PI            80 W        92 W       650 W     Ok        
2        NXA-PAC-650W-PI            75 W        89 W       650 W     Ok        


Power Usage Summary:
--------------------
Power Supply redundancy mode (configured)                PS-Redundant
Power Supply redundancy mode (operational)               PS-Redundant

Total Power Capacity (based on configured mode)             650.00 W
Total Grid-A (first half of PS slots) Power Capacity        650.00 W
Total Grid-B (second half of PS slots) Power Capacity       650.00 W
Total Power of all Inputs (cumulative)                     1300.00 W
Total Power Output (actual draw)                            155.00 W
Total Power Input (actual draw)                             181.00 W
Total Power Allocated (budget)                                N/A   
Total Power Available for additional modules                  N/A   



Temperature:
--------------------------------------------------------------------
Module   Sensor        MajorThresh   MinorThres   CurTemp     Status
                       (Celsius)     (Celsius)    (Celsius)         
--------------------------------------------------------------------
1        FRONT           70              42          26         Ok             
1        BACK            80              70          32         Ok             
1        CPU             90              80          45         Ok             
1        Sugarbowl       100             90          51         Ok
`
	expectedNxos2Metrics = map[string]float64{
		prefix + "fan_operational_status_info{fan=Fan1(sys_fan1),hw=--,model=NXA-FAN-30CFM-B,target=test.test}": 1,
		prefix + "fan_operational_status_info{fan=Fan2(sys_fan2),hw=--,model=NXA-FAN-30CFM-B,target=test.test}": 1,
		prefix + "fan_operational_status_info{fan=Fan3(sys_fan3),hw=--,model=NXA-FAN-30CFM-B,target=test.test}": 1,
		prefix + "fan_operational_status_info{fan=Fan4(sys_fan4),hw=--,model=NXA-FAN-30CFM-B,target=test.test}": 1,
		prefix + "fan_operational_status_info{fan=Fan_in_PS1,hw=--,model=--,target=test.test}":                  1,
		prefix + "fan_operational_status_info{fan=Fan_in_PS2,hw=--,model=--,target=test.test}":                  1,
		prefix + "powersupply_actual_input_watts{model=NXA-PAC-650W-PI,supply=1,target=test.test}":              92,
		prefix + "powersupply_actual_input_watts{model=NXA-PAC-650W-PI,supply=2,target=test.test}":              89,
		prefix + "powersupply_actual_output_watts{model=NXA-PAC-650W-PI,supply=1,target=test.test}":             80,
		prefix + "powersupply_actual_output_watts{model=NXA-PAC-650W-PI,supply=2,target=test.test}":             75,
		prefix + "powersupply_capacity_watts{model=NXA-PAC-650W-PI,supply=1,target=test.test}":                  650,
		prefix + "powersupply_capacity_watts{model=NXA-PAC-650W-PI,supply=2,target=test.test}":                  650,
		prefix + "powersupply_operational_info{input_type=,model=NXA-PAC-650W-PI,ps=1,target=test.test}":        1,
		prefix + "powersupply_operational_info{input_type=,model=NXA-PAC-650W-PI,ps=2,target=test.test}":        1,
		prefix + "powersupply_redundancy_configured_info{target=test.test}":                                     1,
		prefix + "powersupply_redundancy_opererational_info{target=test.test}":                                  1,
		prefix + "powersupply_total_power_input_watts{target=test.test}":                                        181,
		prefix + "powersupply_total_power_output_watts{target=test.test}":                                       155,
		prefix + "powersupply_voltage_volts{target=test.test}":                                                  12,
		prefix + "temperature_current_celsius{module=1,sensor=BACK,target=test.test}":                           32,
		prefix + "temperature_current_celsius{module=1,sensor=CPU,target=test.test}":                            45,
		prefix + "temperature_current_celsius{module=1,sensor=FRONT,target=test.test}":                          26,
		prefix + "temperature_current_celsius{module=1,sensor=Sugarbowl,target=test.test}":                      51,
		prefix + "temperature_major_threshold_celsius{module=1,sensor=BACK,target=test.test}":                   80,
		prefix + "temperature_major_threshold_celsius{module=1,sensor=CPU,target=test.test}":                    90,
		prefix + "temperature_major_threshold_celsius{module=1,sensor=FRONT,target=test.test}":                  70,
		prefix + "temperature_major_threshold_celsius{module=1,sensor=Sugarbowl,target=test.test}":              100,
		prefix + "temperature_minor_threshold_celsius{module=1,sensor=BACK,target=test.test}":                   70,
		prefix + "temperature_minor_threshold_celsius{module=1,sensor=CPU,target=test.test}":                    80,
		prefix + "temperature_minor_threshold_celsius{module=1,sensor=FRONT,target=test.test}":                  42,
		prefix + "temperature_minor_threshold_celsius{module=1,sensor=Sugarbowl,target=test.test}":              90,
	}

	iosXe1 = `
Number of Critical alarms:  0
Number of Major alarms:     0
Number of Minor alarms:     0

Slot    Sensor       Current State       Reading
----    ------       -------------       -------
 P0    PEM Iout         Normal           6 A
 P0    PEM Vout         Normal           12 V DC
 P0    PEM Vin          Normal           53 V AC
 P0    Temp: PEM        Normal           28 Celsius
 P0    Temp: FC         Fan Speed 65%    20 Celsius
 P1    PEM Iout         Normal           6 A
 P1    PEM Vout         Normal           12 V DC
 P1    PEM Vin          Normal           53 V DC
 P1    Temp: PEM        Normal           26 Celsius
 P1    Temp: FC         Fan Speed 65%    20 Celsius
 R0    VCP 1: VX1       Normal           1494 mV
 R0    VCP 1: VX2       Normal           745 mV
 R0    VCP 1: VX3       Normal           1205 mV
 R0    VCP 1: VP1       Normal           5003 mV
 R0    VCP 1: VP2       Normal           3296 mV
 R0    VCP 1: VP3       Normal           2494 mV
 R0    VCP 1: VP4       Normal           1794 mV
 R0    VCP 1: VH        Normal           11945 mV
 R0    VCP 2: VX2       Normal           1048 mV
 R0    VCP 2: VX4       Normal           903 mV
 R0    VCP 2: VX5       Normal           1103 mV
 R0    VCP 2: VP1       Normal           1498 mV
 R0    VCP 2: VP2       Normal           962 mV
 R0    VCP 2: VP3       Normal           1104 mV
 R0    VCP 2: VP4       Normal           1103 mV
 R0    VCP 2: VH        Normal           11961 mV
 R0    Temp: Inlet 1    Normal           20 Celsius
 R0    Temp: Outlet 1   Normal           25 Celsius
 R0    Temp: Octeon     Normal           30 Celsius
 R0    Temp: Outlet 2   Normal           24 Celsius
 R0    Temp: CPU Die    Normal           30 Celsius
 R0    VDP 1: VX1       Normal           1490 mV
 R0    VDP 1: VX4       Normal           928 mV
 R0    VDP 1: VP1       Normal           991 mV
 R0    VDP 1: VP2       Normal           3285 mV
 R0    VDP 1: VP3       Normal           992 mV
 R0    VDP 1: VP4       Normal           1790 mV
 R0    VDP 1: VH        Normal           12029 mV
 R0    VDP 2: VX2       Normal           4990 mV
 R0    VDP 2: VP1       Normal           1489 mV
 R0    VDP 2: VP2       Normal           846 mV
 R0    VDP 2: VP3       Normal           2482 mV
 R0    VDP 2: VP4       Normal           1196 mV
 R0    VDP 2: VH        Normal           12040 mV
 R0    Temp: Inlet 1    Normal           30 Celsius
 R0    Temp: Outlet 1   Normal           27 Celsius
 R0    Temp: WOLV Die   Normal           38 Celsius
 R0    Temp: YODA Die   Normal           40 Celsius
 `

	expectedIosXe1Metrics = map[string]float64{
		prefix + "critical_alarms_total{target=test.test}":                                       0,
		prefix + "current_amps{sensor=pem iout,slot=p0,target=test.test}":                        6,
		prefix + "current_amps{sensor=pem iout,slot=p1,target=test.test}":                        6,
		prefix + "fan_speed_percentage{sensor=temp: fc,slot=p0,taget=test.test}":                 65,
		prefix + "fan_speed_percentage{sensor=temp: fc,slot=p1,taget=test.test}":                 65,
		prefix + "major_alarms_total{target=test.test}":                                          0,
		prefix + "minor_alarms_total{target=test.test}":                                          0,
		prefix + "temperature_current_celsius{module=p0,sensor=temp: fc,target=test.test}":       20,
		prefix + "temperature_current_celsius{module=p0,sensor=temp: pem,target=test.test}":      28,
		prefix + "temperature_current_celsius{module=p1,sensor=temp: fc,target=test.test}":       20,
		prefix + "temperature_current_celsius{module=p1,sensor=temp: pem,target=test.test}":      26,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: cpu die,target=test.test}":  30,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: inlet 1,target=test.test}":  30,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: octeon,target=test.test}":   30,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: outlet 1,target=test.test}": 27,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: outlet 2,target=test.test}": 24,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: wolv die,target=test.test}": 38,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: yoda die,target=test.test}": 40,
		prefix + "voltage_reading_volts{sensor=pem vin,slot=p0,target=test.test}":                53,
		prefix + "voltage_reading_volts{sensor=pem vin,slot=p1,target=test.test}":                53,
		prefix + "voltage_reading_volts{sensor=pem vout,slot=p0,target=test.test}":               12,
		prefix + "voltage_reading_volts{sensor=pem vout,slot=p1,target=test.test}":               12,
		prefix + "voltage_reading_volts{sensor=vcp 1: vh,slot=r0,target=test.test}":              11.945,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp1,slot=r0,target=test.test}":             5.003,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp2,slot=r0,target=test.test}":             3.296,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp3,slot=r0,target=test.test}":             2.494,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp4,slot=r0,target=test.test}":             1.794,
		prefix + "voltage_reading_volts{sensor=vcp 1: vx1,slot=r0,target=test.test}":             1.494,
		prefix + "voltage_reading_volts{sensor=vcp 1: vx2,slot=r0,target=test.test}":             0.745,
		prefix + "voltage_reading_volts{sensor=vcp 1: vx3,slot=r0,target=test.test}":             1.205,
		prefix + "voltage_reading_volts{sensor=vcp 2: vh,slot=r0,target=test.test}":              11.961,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp1,slot=r0,target=test.test}":             1.498,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp2,slot=r0,target=test.test}":             0.962,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp3,slot=r0,target=test.test}":             1.104,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp4,slot=r0,target=test.test}":             1.103,
		prefix + "voltage_reading_volts{sensor=vcp 2: vx2,slot=r0,target=test.test}":             1.048,
		prefix + "voltage_reading_volts{sensor=vcp 2: vx4,slot=r0,target=test.test}":             0.903,
		prefix + "voltage_reading_volts{sensor=vcp 2: vx5,slot=r0,target=test.test}":             1.103,
		prefix + "voltage_reading_volts{sensor=vdp 1: vh,slot=r0,target=test.test}":              12.029,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp1,slot=r0,target=test.test}":             0.991,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp2,slot=r0,target=test.test}":             3.285,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp3,slot=r0,target=test.test}":             0.992,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp4,slot=r0,target=test.test}":             1.79,
		prefix + "voltage_reading_volts{sensor=vdp 1: vx1,slot=r0,target=test.test}":             1.49,
		prefix + "voltage_reading_volts{sensor=vdp 1: vx4,slot=r0,target=test.test}":             0.928,
		prefix + "voltage_reading_volts{sensor=vdp 2: vh,slot=r0,target=test.test}":              12.04,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp1,slot=r0,target=test.test}":             1.489,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp2,slot=r0,target=test.test}":             0.846,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp3,slot=r0,target=test.test}":             2.482,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp4,slot=r0,target=test.test}":             1.196,
		prefix + "voltage_reading_volts{sensor=vdp 2: vx2,slot=r0,target=test.test}":             4.99,
	}

	iosXe2 = `
 Number of Critical alarms:  0
Number of Major alarms:     0
Number of Minor alarms:     0

 Slot        Sensor          Current State   Reading        Threshold(Minor,Major,Critical,Shutdown)
 ----------  --------------  --------------- ------------   ---------------------------------------
 P0          PEM Iout        Normal          7    A      	na
 P0          PEM Vout        Normal          12   V DC   	na
 P0          PEM Vin         Normal          53   V DC   	na
 P0          Temp: PEM       Normal          26   Celsius	(120,140,160,180)(Celsius)
 P0          Temp: FC        Fan Speed 65%   23   Celsius	(31 ,36 ,46 )(Celsius)
 P1          PEM Iout        Normal          7    A      	na
 P1          PEM Vout        Normal          12   V DC   	na
 P1          PEM Vin         Normal          53   V DC   	na
 P1          Temp: PEM       Normal          26   Celsius	(120,140,160,180)(Celsius)
 P1          Temp: FC        Fan Speed 65%   22   Celsius	(31 ,36 ,46 )(Celsius)
 R0          VCP 1: VX1      Normal          1490 mV     	na
 R0          VCP 1: VX2      Normal          743  mV     	na
 R0          VCP 1: VX3      Normal          1198 mV     	na
 R0          VCP 1: VP1      Normal          5029 mV     	na
 R0          VCP 1: VP2      Normal          3279 mV     	na
 R0          VCP 1: VP3      Normal          2493 mV     	na
 R0          VCP 1: VP4      Normal          1791 mV     	na
 R0          VCP 1: VH       Normal          11956mV     	na
 R0          VCP 2: VX2      Normal          1045 mV     	na
 R0          VCP 2: VX4      Normal          900  mV     	na
 R0          VCP 2: VX5      Normal          1110 mV     	na
 R0          VCP 2: VP1      Normal          1493 mV     	na
 R0          VCP 2: VP2      Normal          943  mV     	na
 R0          VCP 2: VP3      Normal          1104 mV     	na
 R0          VCP 2: VP4      Normal          1098 mV     	na
 R0          VCP 2: VH       Normal          11961mV     	na
 R0          Temp: Inlet 1   Normal          22   Celsius	(45 ,55 ,65 ,70 )(Celsius)
 R0          Temp: Outlet 1  Normal          26   Celsius	(55 ,65 ,75 ,80 )(Celsius)
 R0          Temp: Octeon    Normal          33   Celsius	(60 ,70 ,80 ,85 )(Celsius)
 R0          Temp: Outlet 2  Normal          25   Celsius	(55 ,65 ,75 ,80 )(Celsius)
 R0          Temp: CPU Die   Normal          30   Celsius	(65 ,75 ,79 ,83 )(Celsius)
 R0          VDP 1: VX1      Normal          1488 mV     	na
 R0          VDP 1: VX4      Normal          929  mV     	na
 R0          VDP 1: VP1      Normal          992  mV     	na
 R0          VDP 1: VP2      Normal          3292 mV     	na
 R0          VDP 1: VP3      Normal          996  mV     	na
 R0          VDP 1: VP4      Normal          1792 mV     	na
 R0          VDP 1: VH       Normal          12045mV     	na
 R0          VDP 2: VX2      Normal          4980 mV     	na
 R0          VDP 2: VP1      Normal          1490 mV     	na
 R0          VDP 2: VP2      Normal          843  mV     	na
 R0          VDP 2: VP3      Normal          2473 mV     	na
 R0          VDP 2: VP4      Normal          1199 mV     	na
 R0          VDP 2: VH       Normal          12029mV     	na
 R0          Temp: Inlet 1   Normal          32   Celsius	(55 ,65 ,75 ,80 )(Celsius)
 R0          Temp: Outlet 1  Normal          28   Celsius	(55 ,65 ,75 ,80 )(Celsius)
 R0          Temp: WOLV Die  Normal          41   Celsius	(75 ,85 ,95 ,100)(Celsius)
 R0          Temp: YODA Die  Normal          42   Celsius	(90 ,100,110,120)(Celsius)
 `
	expectedIosXe2Metrics = map[string]float64{
		prefix + "critical_alarms_total{target=test.test}":                                                  0,
		prefix + "current_amps{sensor=pem iout,slot=p0,target=test.test}":                                   7,
		prefix + "current_amps{sensor=pem iout,slot=p1,target=test.test}":                                   7,
		prefix + "fan_speed_percentage{sensor=temp: fc,slot=p0,taget=test.test}":                            65,
		prefix + "fan_speed_percentage{sensor=temp: fc,slot=p1,taget=test.test}":                            65,
		prefix + "major_alarms_total{target=test.test}":                                                     0,
		prefix + "minor_alarms_total{target=test.test}":                                                     0,
		prefix + "temperature_critical_threshold_celsius{module=p0,sensor=temp: fc,target=test.test}":       46,
		prefix + "temperature_critical_threshold_celsius{module=p0,sensor=temp: pem,target=test.test}":      160,
		prefix + "temperature_critical_threshold_celsius{module=p1,sensor=temp: fc,target=test.test}":       46,
		prefix + "temperature_critical_threshold_celsius{module=p1,sensor=temp: pem,target=test.test}":      160,
		prefix + "temperature_critical_threshold_celsius{module=r0,sensor=temp: cpu die,target=test.test}":  79,
		prefix + "temperature_critical_threshold_celsius{module=r0,sensor=temp: inlet 1,target=test.test}":  75,
		prefix + "temperature_critical_threshold_celsius{module=r0,sensor=temp: octeon,target=test.test}":   80,
		prefix + "temperature_critical_threshold_celsius{module=r0,sensor=temp: outlet 1,target=test.test}": 75,
		prefix + "temperature_critical_threshold_celsius{module=r0,sensor=temp: outlet 2,target=test.test}": 75,
		prefix + "temperature_critical_threshold_celsius{module=r0,sensor=temp: wolv die,target=test.test}": 95,
		prefix + "temperature_critical_threshold_celsius{module=r0,sensor=temp: yoda die,target=test.test}": 110,
		prefix + "temperature_current_celsius{module=p0,sensor=temp: fc,target=test.test}":                  23,
		prefix + "temperature_current_celsius{module=p0,sensor=temp: pem,target=test.test}":                 26,
		prefix + "temperature_current_celsius{module=p1,sensor=temp: fc,target=test.test}":                  22,
		prefix + "temperature_current_celsius{module=p1,sensor=temp: pem,target=test.test}":                 26,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: cpu die,target=test.test}":             30,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: inlet 1,target=test.test}":             32,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: octeon,target=test.test}":              33,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: outlet 1,target=test.test}":            28,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: outlet 2,target=test.test}":            25,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: wolv die,target=test.test}":            41,
		prefix + "temperature_current_celsius{module=r0,sensor=temp: yoda die,target=test.test}":            42,
		prefix + "temperature_major_threshold_celsius{module=p0,sensor=temp: fc,target=test.test}":          36,
		prefix + "temperature_major_threshold_celsius{module=p0,sensor=temp: pem,target=test.test}":         140,
		prefix + "temperature_major_threshold_celsius{module=p1,sensor=temp: fc,target=test.test}":          36,
		prefix + "temperature_major_threshold_celsius{module=p1,sensor=temp: pem,target=test.test}":         140,
		prefix + "temperature_major_threshold_celsius{module=r0,sensor=temp: cpu die,target=test.test}":     75,
		prefix + "temperature_major_threshold_celsius{module=r0,sensor=temp: inlet 1,target=test.test}":     65,
		prefix + "temperature_major_threshold_celsius{module=r0,sensor=temp: octeon,target=test.test}":      70,
		prefix + "temperature_major_threshold_celsius{module=r0,sensor=temp: outlet 1,target=test.test}":    65,
		prefix + "temperature_major_threshold_celsius{module=r0,sensor=temp: outlet 2,target=test.test}":    65,
		prefix + "temperature_major_threshold_celsius{module=r0,sensor=temp: wolv die,target=test.test}":    85,
		prefix + "temperature_major_threshold_celsius{module=r0,sensor=temp: yoda die,target=test.test}":    100,
		prefix + "temperature_minor_threshold_celsius{module=p0,sensor=temp: fc,target=test.test}":          31,
		prefix + "temperature_minor_threshold_celsius{module=p0,sensor=temp: pem,target=test.test}":         120,
		prefix + "temperature_minor_threshold_celsius{module=p1,sensor=temp: fc,target=test.test}":          31,
		prefix + "temperature_minor_threshold_celsius{module=p1,sensor=temp: pem,target=test.test}":         120,
		prefix + "temperature_minor_threshold_celsius{module=r0,sensor=temp: cpu die,target=test.test}":     65,
		prefix + "temperature_minor_threshold_celsius{module=r0,sensor=temp: inlet 1,target=test.test}":     55,
		prefix + "temperature_minor_threshold_celsius{module=r0,sensor=temp: octeon,target=test.test}":      60,
		prefix + "temperature_minor_threshold_celsius{module=r0,sensor=temp: outlet 1,target=test.test}":    55,
		prefix + "temperature_minor_threshold_celsius{module=r0,sensor=temp: outlet 2,target=test.test}":    55,
		prefix + "temperature_minor_threshold_celsius{module=r0,sensor=temp: wolv die,target=test.test}":    75,
		prefix + "temperature_minor_threshold_celsius{module=r0,sensor=temp: yoda die,target=test.test}":    90,
		prefix + "temperature_shutdown_threshold_celsius{module=p0,sensor=temp: pem,target=test.test}":      180,
		prefix + "temperature_shutdown_threshold_celsius{module=p1,sensor=temp: pem,target=test.test}":      180,
		prefix + "temperature_shutdown_threshold_celsius{module=r0,sensor=temp: cpu die,target=test.test}":  83,
		prefix + "temperature_shutdown_threshold_celsius{module=r0,sensor=temp: inlet 1,target=test.test}":  80,
		prefix + "temperature_shutdown_threshold_celsius{module=r0,sensor=temp: octeon,target=test.test}":   85,
		prefix + "temperature_shutdown_threshold_celsius{module=r0,sensor=temp: outlet 1,target=test.test}": 80,
		prefix + "temperature_shutdown_threshold_celsius{module=r0,sensor=temp: outlet 2,target=test.test}": 80,
		prefix + "temperature_shutdown_threshold_celsius{module=r0,sensor=temp: wolv die,target=test.test}": 100,
		prefix + "temperature_shutdown_threshold_celsius{module=r0,sensor=temp: yoda die,target=test.test}": 120,
		prefix + "voltage_reading_volts{sensor=pem vin,slot=p0,target=test.test}":                           53,
		prefix + "voltage_reading_volts{sensor=pem vin,slot=p1,target=test.test}":                           53,
		prefix + "voltage_reading_volts{sensor=pem vout,slot=p0,target=test.test}":                          12,
		prefix + "voltage_reading_volts{sensor=pem vout,slot=p1,target=test.test}":                          12,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp1,slot=r0,target=test.test}":                        5.029,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp2,slot=r0,target=test.test}":                        3.279,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp3,slot=r0,target=test.test}":                        2.493,
		prefix + "voltage_reading_volts{sensor=vcp 1: vp4,slot=r0,target=test.test}":                        1.791,
		prefix + "voltage_reading_volts{sensor=vcp 1: vx1,slot=r0,target=test.test}":                        1.49,
		prefix + "voltage_reading_volts{sensor=vcp 1: vx2,slot=r0,target=test.test}":                        0.743,
		prefix + "voltage_reading_volts{sensor=vcp 1: vx3,slot=r0,target=test.test}":                        1.198,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp1,slot=r0,target=test.test}":                        1.493,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp2,slot=r0,target=test.test}":                        0.943,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp3,slot=r0,target=test.test}":                        1.104,
		prefix + "voltage_reading_volts{sensor=vcp 2: vp4,slot=r0,target=test.test}":                        1.098,
		prefix + "voltage_reading_volts{sensor=vcp 2: vx2,slot=r0,target=test.test}":                        1.045,
		prefix + "voltage_reading_volts{sensor=vcp 2: vx4,slot=r0,target=test.test}":                        0.9,
		prefix + "voltage_reading_volts{sensor=vcp 2: vx5,slot=r0,target=test.test}":                        1.11,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp1,slot=r0,target=test.test}":                        0.992,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp2,slot=r0,target=test.test}":                        3.292,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp3,slot=r0,target=test.test}":                        0.996,
		prefix + "voltage_reading_volts{sensor=vdp 1: vp4,slot=r0,target=test.test}":                        1.792,
		prefix + "voltage_reading_volts{sensor=vdp 1: vx1,slot=r0,target=test.test}":                        1.488,
		prefix + "voltage_reading_volts{sensor=vdp 1: vx4,slot=r0,target=test.test}":                        0.929,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp1,slot=r0,target=test.test}":                        1.49,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp2,slot=r0,target=test.test}":                        0.843,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp3,slot=r0,target=test.test}":                        2.473,
		prefix + "voltage_reading_volts{sensor=vdp 2: vp4,slot=r0,target=test.test}":                        1.199,
		prefix + "voltage_reading_volts{sensor=vdp 2: vx2,slot=r0,target=test.test}":                        4.98,
	}

	ios1 = `
FAN in PS-1 is OK
FAN in PS-2 is OK
SYSTEM TEMPERATURE is GREEN
SYSTEM Temperature Value: 42.5 Degree Celsius
SYSTEM Temperature State: GREEN
SYSTEM Low Temperature Alert Threshold: 0.0 Degree Celsius
SYSTEM Low Temperature Shutdown Threshold: -20.0 Degree Celsius
SYSTEM High Temperature Alert Threshold: 58.0 Degree Celsius
SYSTEM High Temperature Shutdown Threshold: 80.0 Degree Celsius
POWER SUPPLY 1 Temperature Value: 39.7500 Degree Celsius
POWER SUPPLY 1 Temperature Alert Threshold: 85.0000 Degree Celsius
POWER SUPPLY 1 Temperature Shutdown Threshold: 110.0000 Degree Celsius
POWER SUPPLY 2 Temperature Value: 40.2500 Degree Celsius
POWER SUPPLY 2 Temperature Alert Threshold: 85.0000 Degree Celsius
POWER SUPPLY 2 Temperature Shutdown Threshold: 110.0000 Degree Celsius
POWER SUPPLY 1 is DC OK
POWER SUPPLY 2 is DC OK

ALARM CONTACT 1 is not asserted
ALARM CONTACT 2 is not asserted
ALARM CONTACT 3 is not asserted
ALARM CONTACT 4 is not asserted
`

	expectedIos1Metrics = map[string]float64{
		prefix + "alarm_contacted_asserted_info{contact=1,target=test.test}":                           0,
		prefix + "alarm_contacted_asserted_info{contact=2,target=test.test}":                           0,
		prefix + "alarm_contacted_asserted_info{contact=3,target=test.test}":                           0,
		prefix + "alarm_contacted_asserted_info{contact=4,target=test.test}":                           0,
		prefix + "fan_operational_status_info{fan= ps-1,hw=,model=,target=test.test}":                  1,
		prefix + "fan_operational_status_info{fan= ps-2,hw=,model=,target=test.test}":                  1,
		prefix + "powersupply_operational_info{input_type=,model=,ps=power supply 1,target=test.test}": 1,
		prefix + "powersupply_operational_info{input_type=,model=,ps=power supply 2,target=test.test}": 1,
		prefix + "system_temperature_status_info{status=green,target=test.test}":                       1,
        prefix + "temperature_current_celsius{module=,sensor=power supply 1,target=test.test}": 39.75,
        prefix + "temperature_current_celsius{module=,sensor=power supply 2,target=test.test}": 40.25,
		prefix + "temperature_high_alarm_threshold_celsius{sensor=power supply 1,target=test.test}":    85,
		prefix + "temperature_high_alarm_threshold_celsius{sensor=power supply 2,target=test.test}":    85,
		prefix + "temperature_high_alarm_threshold_celsius{sensor=system,target=test.test}":            58,
		prefix + "temperature_high_shutdown_threshold_celsius{sensor=power supply 1,target=test.test}": 110,
		prefix + "temperature_high_shutdown_threshold_celsius{sensor=power supply 2,target=test.test}": 110,
		prefix + "temperature_high_shutdown_threshold_celsius{sensor=system,target=test.test}":         80,
		prefix + "temperature_low_alarm_threshold_celsius{sensor=system,target=test.test}":             0,
		prefix + "temperature_low_shutdown_threshold_celsius{sensor=system,target=test.test}":          -20,
	}

	ios2 = `
FAN in PS-1 is OK
FAN in PS-2 is OK
SYSTEM TEMPERATURE is GREEN
SYSTEM Temperature Value: 49.5 Degree Celsius
SYSTEM Temperature State: GREEN
SYSTEM Low Temperature Alert Threshold: 0.0 Degree Celsius
SYSTEM Low Temperature Shutdown Threshold: -20.0 Degree Celsius
SYSTEM High Temperature Alert Threshold: 58.0 Degree Celsius
SYSTEM High Temperature Shutdown Threshold: 80.0 Degree Celsius
POWER SUPPLY 1 Temperature Value: 49.5000 Degree Celsius
POWER SUPPLY 1 Temperature Alert Threshold: 85.0000 Degree Celsius
POWER SUPPLY 1 Temperature Shutdown Threshold: 110.0000 Degree Celsius
POWER SUPPLY 2 Temperature Value: 47.7500 Degree Celsius
POWER SUPPLY 2 Temperature Alert Threshold: 85.0000 Degree Celsius
POWER SUPPLY 2 Temperature Shutdown Threshold: 110.0000 Degree Celsius
POWER SUPPLY 1 is DC OK
POWER SUPPLY 2 is DC OK

ALARM CONTACT 1 is not asserted
ALARM CONTACT 2 is not asserted
ALARM CONTACT 3 is not asserted
ALARM CONTACT 4 is not asserted
`
	expectedIos2Metrics = map[string]float64{
		prefix + "alarm_contacted_asserted_info{contact=1,target=test.test}":                           0,
		prefix + "alarm_contacted_asserted_info{contact=2,target=test.test}":                           0,
		prefix + "alarm_contacted_asserted_info{contact=3,target=test.test}":                           0,
		prefix + "alarm_contacted_asserted_info{contact=4,target=test.test}":                           0,
		prefix + "fan_operational_status_info{fan= ps-1,hw=,model=,target=test.test}":                  1,
		prefix + "fan_operational_status_info{fan= ps-2,hw=,model=,target=test.test}":                  1,
		prefix + "powersupply_operational_info{input_type=,model=,ps=power supply 1,target=test.test}": 1,
		prefix + "powersupply_operational_info{input_type=,model=,ps=power supply 2,target=test.test}": 1,
		prefix + "system_temperature_status_info{status=green,target=test.test}":                       1,
        prefix + "temperature_current_celsius{module=,sensor=power supply 2,target=test.test}": 47.75,
        prefix + "temperature_current_celsius{module=,sensor=power supply 1,target=test.test}": 49.5,
		prefix + "temperature_high_alarm_threshold_celsius{sensor=power supply 1,target=test.test}":    85,
		prefix + "temperature_high_alarm_threshold_celsius{sensor=power supply 2,target=test.test}":    85,
		prefix + "temperature_high_alarm_threshold_celsius{sensor=system,target=test.test}":            58,
		prefix + "temperature_high_shutdown_threshold_celsius{sensor=power supply 1,target=test.test}": 110,
		prefix + "temperature_high_shutdown_threshold_celsius{sensor=power supply 2,target=test.test}": 110,
		prefix + "temperature_high_shutdown_threshold_celsius{sensor=system,target=test.test}":         80,
		prefix + "temperature_low_alarm_threshold_celsius{sensor=system,target=test.test}":             0,
		prefix + "temperature_low_shutdown_threshold_celsius{sensor=system,target=test.test}":          -20,
	}
)
