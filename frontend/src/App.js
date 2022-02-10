import React, { useEffect, useState } from 'react';
import './App.css';
import 'antd/dist/antd.css';

import { 
  Typography,
  Button,
  Spin,
  Divider,
  Space,
  Form,
  Slider,
  Tooltip,
  message,
  Switch,
} from 'antd';

import {
  DownOutlined,
  UpOutlined,
} from '@ant-design/icons';

const { Title, Text } = Typography;


function WiggleOptions({ isReady, isWiggling, wiggleDuration, waitTime, setWiggleDuration, setWaitTime }) {
  const [form] = Form.useForm();

  const onSubmit = async ({wiggleSpeed, waitTime}) => {
    // Wait for ready signal
    if (!isReady) return;

    // Send config to backend
    await window.go.main.Wiggler.SetConfig(wiggleSpeed, waitTime).then(() => {
      message.success("Updated Wiggle Config!");
    }).catch(err => {
      message.error("Error Setting Wiggle Config");
      console.error(err);
    });

    // Restart (if already started)
    if (isWiggling) {
      await window.go.main.Wiggler.StartWiggle();
    }

    // Get config from backend
    const newWiggleSpeed = await window.go.main.Wiggler.GetMoveSpeed();
    const newWaitTime = await window.go.main.Wiggler.GetWaitTime();

    form.setFieldsValue({
      wiggleSpeed: newWiggleSpeed,
      waitTime: newWaitTime
    });

    // Set the state data too
    setWiggleDuration(newWiggleSpeed);
    setWaitTime(newWaitTime);

  };
  const onReset = () => {
    form.resetFields();
  };


  return (
    <div className="wiggle-options">
      <Divider />
      <Title level={4}>
        Wiggler Options
      </Title>
      <Form
        name="wiggle-options"
        form={form}
        onFinish={onSubmit}
        initialValues={{
          wiggleSpeed: wiggleDuration,
          waitTime,
        }}
      >
        <Form.Item 
          label="Wiggle for"
          name="wiggleSpeed"
        >
          <Slider 
            min={0.1}
            max={10}
            step={1}
            tipFormatter={(value) => `${value} second${ value === 1 ? "" : "s" }`}
            tooltipVisible
          />
        </Form.Item>
        <div style={{ height: "10px", }} />
        <Form.Item 
          label="Before wiggling again, wait"
          name="waitTime"
        >
          <Slider 
            min={0}
            max={60}
            step={1}
            tipFormatter={(value) => `${value} second${ value === 1 ? "" : "s" }`}
            tooltipVisible
          />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button 
              type="primary"
              htmlType="submit"
              disabled={!isReady}
            >
              Update { isWiggling && "and Restart" }
            </Button>
            <Button 
              type="default" 
              disabled={!isReady}
              onClick={onReset}
            >
              Reset
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}


function App() {
  const [isReady, setIsReady] = useState(true);
  const [optionsVisible, setOptionsVisible] = useState(false);

  const [isWiggling, setIsWiggling] = useState(false);
  const [wiggleDuration, setWiggleDuration] = useState(1);
  const [waitTime, setWaitTime] = useState(2);

  useEffect(() => {
    document.addEventListener("keydown", async (e) => {
      // If the "Escape" key is pressed and a wiggle event is occuring...
      if (!isWiggling) return;
      if (e.key !== "Escape" || e.key !== " " || e.key !== "Enter" || e.key !== "q") return;

      // Stop the wiggle
      await window.go.main.Wiggler.StopWiggle()

      // Set the state
      const wiggleState = await window.go.main.Wiggler.IsWiggling()
      await setIsWiggling(wiggleState);
    });
  }, []);

  // Register Wails event trigger functions
  useEffect(() => {
    // Run when wiggler config is confirmed
    window.runtime.EventsOn("config-set", ({duration, time}) => {
      console.log("config-set", duration, time);
      // setWiggleDuration(duration);
      // setWaitTime(time);
    });

    // Run when a wiggler START event is confirmed
    window.runtime.EventsOn("wiggle-started", () => {
      setIsWiggling(true);
    });

    // Run when a wiggler STOP event is confirmed
    window.runtime.EventsOn("wiggle-stopped", () => {
      setIsWiggling(true);
    });

    // Run when the backend is stopping
    window.runtime.EventsOn("stopped", () => {
      setIsReady(false);
      setIsWiggling(false);
    });
  }, []);

  return (
    <div 
      id="app" 
      className="App"
      style={{
        maxWidth: "600px",
        margin: "0 auto",
        padding: "20px",
      }}
    >
      <Title>The Wiggler</Title>
      <Space direction="vertical">
        <Text>
          Click the button to start the wiggler. Click it again to stop.
        </Text>
        <Text>
          To configure the wiggler, click the "More Options" button.
        </Text>
      </Space>

      <Divider />

      <Space direction="horizontal" size="large">
        <Space>
          <Button
            type={isWiggling ? "ghost" : "primary"}
            disabled={!isReady}
            onClick={async () => {
              if (isWiggling) {
                await window.go.main.Wiggler.StopWiggle();
              } else {
                await window.go.main.Wiggler.StartWiggle();
              }
              const wiggleState = await window.go.main.Wiggler.IsWiggling();
              setIsWiggling(wiggleState);
            }}
          >
            { isWiggling ? "Stop" : "Start" } Wiggling
          </Button>
          <Tooltip title="Display More Options">
            <Button
              type="secondary"
              icon={optionsVisible ? <UpOutlined /> : <DownOutlined />}
              onClick={async () => {
                if (optionsVisible) {
                  await window.go.main.Wiggler.SetWindowSmall();
                } else {
                  await window.go.main.Wiggler.SetWindowLarge();
                }
                setOptionsVisible(!optionsVisible)
              }}
            />
          </Tooltip>
        </Space>
        { isWiggling && <Spin /> }
      </Space>

      {optionsVisible && (
        <WiggleOptions 
          isWiggling={isWiggling}
          wiggleDuration={wiggleDuration}
          waitTime={waitTime}
          isReady={isReady}
          setWiggleDuration={setWiggleDuration}
          setWaitTime={setWaitTime}
        /> 
      )}

    </div>
  );
}

export default App;
