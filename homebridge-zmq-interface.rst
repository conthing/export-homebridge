homebridge-zmq协议编写
##########################

运行流程
**************

    .. image:: process.png
        :width: 698px

通过command zmq发送的命令
****************************
init命令
=============
发送QRcode和commandport

    .. code:: json
	
       {
         "command": {
            "name": "init",
            "params": {
               "QRcode": "X-HM://0023JY2ILGN51",
               "commandport":51826
		    }
	     }
       }


灯光对应命令
===============
开关灯

    .. code:: json
    
       {
         "name":"客厅灯",
         "id":"1",
         "service":"LightBulb",
         "command": {
            "name": "setLightBulbState",
            "params": {
               "onOrOff":true
		    }
	     }
       }

设置灯光亮度

    .. code:: json
    
       {
         "name":"客厅灯",
         "id":"1",
         "service":"LightBulb",
         "command": {
            "name": "setLightBulbState",
            "params": {
               "brightness":88
		    }
	     }
       }


窗帘对应命令
==============
设置窗帘目标行程

    .. code:: json
    
       {
         "name":"客厅窗帘",
         "id":"2",
         "service":"WindowCovering",
         "command": {
            "name": "setWindowCoveringTargetPosition",
            "params": {
               "percent":100
		    }
	     }
       }
   
设置窗帘目标角度

    .. code:: json
    
       {
         "name":"客厅窗帘",
         "id":"2",
         "service":"WindowCovering",
         "command": {
            "name": "setWindowCoveringTargetHorizontalTiltAngle",
            "params": {
               "angle":0
		    }
	     }
       }

空调对应命令
===================
设置空调模式及目标温度

    .. code:: json
    
       {
         "name":"客厅空调",
         "id":"5",
         "service":"Thermostat",
         "command": {
            "name": "setTargetTemperature",
            "params": {
               "mode":"AC",
               "t_target":26
		    }
	     }
       }

设置空调目标模式

    .. code:: json
    
       {
         "name":"客厅空调",
         "id":"5",
         "service":"Thermostat",
         "command": {
            "name": "setTargetHeatingCoolingState",
            "params": {
               "mode":"AC"
		    }
	     }
       }

设置空调温度的单位

    .. code:: json
    
       {
         "name":"客厅空调",
         "id":"5",
         "service":"Thermostat",
         "command": {
            "name": "setTemperatureDisplayUnits",
            "params": {
               "unit":"C"
		    }
	     }
       }

传感器对应命令
=================
窗户

    .. code:: json
    
       {
         "name":"客厅窗户",
         "id":"9",
         "service":"Window",
         "command": {
            "name": "setWindowTargetPosition",
            "params": {
               "state":1
		    }
	     }
       }
   
门

    .. code:: json
    
       {
         "name":"客厅门",
         "id":"12",
         "service":"Door",
         "command": {
            "name": "setDoorTargetPosition",
            "params": {
               "state":1
		    }
	     }
       }

门锁

    .. code:: json
    
       {
         "name":"客厅门锁",
         "id":"15",
         "service":"Lock",
         "command": {
            "name": "setLockTargetState",
            "params": {
               "state":1
		    }
	     }
       }

开关

    .. code:: json
    
       {
         "name":"客厅开关",
         "id":"16",
         "service":"Switch",
         "command": {
            "name": "setSwitchState",
            "params": {
               "state":1
		    }
	     }
       }

风扇

    .. code:: json
    
       {
         "name":"客厅风扇",
         "id":"20",
         "service":"Fan",
         "command": {
            "name": "setFanState",
            "params": {
               "state":1
		    }
	     }
       }

    .. code:: json
    
       {
         "name":"新风风扇",
         "id":"23",
         "service":"Fanv2",
         "command": {
            "name": "setRotationDirection",
            "params": {
               "fan1vol":40
		    }
	     }
       }


通过status zmq发送的命令
****************************
所有设备状态的变化
=====================

    .. code:: json
    
        {
          "status": [{
            "id": "1",
            "name": "客厅灯",
            "service": "LightBulb",
            "characteristic": {
                "on": true,
                "brightness": 100
            }
         },
         {
            "id": "2",
            "name": "客厅窗帘",
            "service": "WindowCovering",
            "characteristic": {
                "percent": 100
            }
         }
        ]}

灯
==========
灯亮度

    .. code:: json
    
        {
          "status": [{
            "id": "1",
            "name": "客厅灯",
            "service": "LightBulb",
            "characteristic": {
                "on": true,
                "brightness": 100
            }
         }]}

窗帘
=========
窗帘行程

    .. code:: json
    
        {
          "status": [{
            "id": "2",
            "name": "客厅窗帘",
            "service": "WindowCovering",
            "characteristic": {
                "percent": 100
            }
         }
		]}

窗帘角度

    .. code:: json
    
        {
          "status": [{
            "id": "2",
            "name": "客厅窗帘",
            "service": "WindowCovering",
            "characteristic": {
                "angle": 0
            }
         }
		]}

空调
========
空调目标模式

    .. code:: json
    
        {
          "status": [{
            "id": "5",
            "name": "客厅空调",
            "service": "Thermostat",
            "characteristic": {
                "target_mode": "AC"
            }
         }
		]}

空调目标温度

    .. code:: json
    
        {
          "status": [{
            "id": "5",
            "name": "客厅空调",
            "service": "Thermostat",
            "characteristic": {
                "t_target": 26
            }
         }
		]}

空调现有温度

    .. code:: json
    
        {
          "status": [{
            "id": "5",
            "name": "客厅空调",
            "service": "Thermostat",
            "characteristic": {
                "t_indoor": 22
            }
         }
		]}

空调现在模式

    .. code:: json
    
        {
          "status": [{
            "id": "5",
            "name": "客厅空调",
            "service": "Thermostat",
            "characteristic": {
                "mode": "AC"
            }
         }
		]}

空调温度单位

    .. code:: json
    
        {
          "status": [{
            "id": "5",
            "name": "客厅空调",
            "service": "Thermostat",
            "characteristic": {
                "unit": "C"
            }
         }
		]}

传感器
=========
窗户

    .. code:: json
    
        {
          "status": [{
            "id": "9",
            "name": "客厅窗户",
            "service": "Window",
            "characteristic": {
                "state": 1
            }
         }
		]}

门

    .. code:: json
    
        {
          "status": [{
            "id": "12",
            "name": "客厅门",
            "service": "Door",
            "characteristic": {
                "state": 1
            }
         }
		]}

门锁

    .. code:: json
    
        {
          "status": [{
            "id": "15",
            "name": "客厅门锁",
            "service": "Lock",
            "characteristic": {
                "state": 1
            }
         }
		]}

开关

    .. code:: json
    
        {
          "status": [{
            "id": "16",
            "name": "客厅开关",
            "service": "Switch",
            "characteristic": {
                "state": 1
            }
         }
		]}

风扇

    .. code:: json
    
        {
          "status": [{
            "id": "20",
            "name": "客厅风扇",
            "service": "Fan",
            "characteristic": {
                "state": 1
            }
         }
		]}

    .. code:: json
    
        {
          "status": [{
            "id": "23",
            "name": "新风风扇",
            "service": "Fanv2",
            "characteristic": {
                "fan1vol": 40
            }
         }
		]}

    .. code:: json
    
        {
          "status": [{
            "id": "23",
            "name": "新风风扇",
            "service": "Fanv2",
            "characteristic": {
                "fan_mode": 1
            }
         }
		]}

空气质量传感器

    .. code:: json
    
        {
          "status": [{
            "id": "22",
            "name": "客厅空气质量传感器",
            "service": "AirQualitySensor",
            "characteristic": {
                "indoor_pm2d5": 50
            }
         }
		]}

    .. code:: json
    
        {
          "status": [{
            "id": "22",
            "name": "客厅空气质量传感器",
            "service": "AirQualitySensor",
            "characteristic": {
                "indoor_co2": 600
            }
         }
		]}

    .. code:: json
    
        {
          "status": [{
            "id": "22",
            "name": "客厅空气质量传感器",
            "service": "AirQualitySensor",
            "characteristic": {
                "indoor_co": 0
            }
         }
		]}

温度传感器

    .. code:: json
    
        {
          "status": [{
            "id": "24",
            "name": "温度传感器",
            "service": "TemperatureSensor",
            "characteristic": {
                "temp_indoor": 26
            }
         }
		]}

湿度传感器

    .. code:: json
    
        {
          "status": [{
            "id": "26",
            "name": "温度传感器",
            "service": "HumiditySensor",
            "characteristic": {
                "humidity_indoor": 50
            }
         }
		]}
